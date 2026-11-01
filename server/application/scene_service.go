package application

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
)

var (
	// ErrNoDriver is returned when a device uses an integration that has no registered driver.
	ErrNoDriver = errors.New("no driver registered for integration")
	// ErrNoActiveScene is returned when an operation requires an active scene.
	ErrNoActiveScene = errors.New("no active scene")
)

// SceneConfigStore reads authored scene configuration. Backed by file storage.
type SceneConfigStore interface {
	GetSceneConfig(ctx context.Context, sceneID string) (*scenes.Config, error)
	// ListSceneConfigs returns all configured scenes, sorted by ID.
	ListSceneConfigs(ctx context.Context) ([]*scenes.Config, error)
}

// RoomConfigStore reads the configured room (devices and target bindings). Backed by file storage.
type RoomConfigStore interface {
	CurrentRoom(ctx context.Context) (*devices.Room, error)
}

// SceneStateStore persists runtime scene state. Backed by the database.
type SceneStateStore interface {
	// LoadSceneState returns the saved state of a scene, or [scenes.ErrStateNotFound] if it was never saved.
	LoadSceneState(ctx context.Context, sceneID string) (*scenes.State, error)
	SaveSceneState(ctx context.Context, state *scenes.State) error
}

// OutputDriver delivers instructions to the devices of one integration (hue, spotify, mqtt, the
// built-in screen client, ...). Implementations live in infrastructure.
type OutputDriver interface {
	// Integration returns the integration identifier this driver handles, see devices.Integration*.
	Integration() string
	// Apply instructs a single device. The instruction kind always matches the device kind. scene is the
	// scene the instruction comes from, e.g. to resolve scene relative asset paths.
	Apply(ctx context.Context, scene scenes.Ref, device *devices.Device, instruction scenes.Instruction) error
}

// SceneService loads, switches and exits scenes and instructs the devices in the room.
//
// Configuration is only ever read. The active scene's state is saved when the scene is loaded and
// when it is exited (by activating another scene or calling [SceneService.Exit]).
type SceneService struct {
	configs SceneConfigStore
	states  SceneStateStore
	rooms   RoomConfigStore
	drivers map[string]OutputDriver

	lock   sync.Mutex
	active *scenes.Scene
}

// NewSceneService creates a new SceneService with the given drivers.
func NewSceneService(configs SceneConfigStore, states SceneStateStore, rooms RoomConfigStore, drivers ...OutputDriver) *SceneService {
	s := &SceneService{
		configs: configs,
		states:  states,
		rooms:   rooms,
		drivers: make(map[string]OutputDriver, len(drivers)),
	}
	for _, d := range drivers {
		s.drivers[d.Integration()] = d
	}

	return s
}

// Active returns the active scene and variant, zero when no scene is active.
func (s *SceneService) Active() scenes.Ref {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.active == nil {
		return scenes.Ref{}
	}

	return s.active.Ref()
}

// Activate loads a scene and instructs the devices. The state of the scene being exited is saved
// first; if that fails the switch is aborted. The loaded scene resumes from its saved state, or
// starts fresh from its configuration. A non-empty variantID switches variant after loading.
// Activating the already active scene only switches variant (or does nothing for an empty variantID).
func (s *SceneService) Activate(ctx context.Context, sceneID, variantID string) (*scenes.Transition, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.active != nil && s.active.ID() == sceneID {
		return s.change(ctx, func(sc *scenes.Scene) error {
			if variantID == "" {
				return nil
			}

			return sc.SetVariant(variantID)
		})
	}

	next, err := s.load(ctx, sceneID)
	if err != nil {
		return nil, err
	}
	if variantID != "" {
		if err := next.SetVariant(variantID); err != nil {
			return nil, err
		}
	}
	if err := s.exit(ctx); err != nil {
		return nil, err
	}
	if err := s.states.SaveSceneState(ctx, next.State); err != nil {
		return nil, fmt.Errorf("saving state of scene %q: %w", sceneID, err)
	}

	return s.transition(ctx, scenes.Ref{}, scenes.Instructions{}, next)
}

// SetVariant switches the variant of the active scene (empty for the default).
func (s *SceneService) SetVariant(ctx context.Context, variantID string) (*scenes.Transition, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.change(ctx, func(sc *scenes.Scene) error { return sc.SetVariant(variantID) })
}

// Adjust applies live instruction overrides to the active scene, e.g. the GM changing the volume.
func (s *SceneService) Adjust(ctx context.Context, overrides scenes.Instructions) (*scenes.Transition, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.change(ctx, func(sc *scenes.Scene) error {
		sc.Adjust(overrides)

		return nil
	})
}

// Scenes returns the configuration of every scene that can be activated.
func (s *SceneService) Scenes(ctx context.Context) ([]*scenes.Config, error) {
	return s.configs.ListSceneConfigs(ctx)
}

// ActiveConfig returns the configuration and the reference of the active scene, nil when no scene is
// active. The configuration is shared and must not be modified.
func (s *SceneService) ActiveConfig() (*scenes.Config, scenes.Ref) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.active == nil {
		return nil, scenes.Ref{}
	}

	return s.active.Config, s.active.Ref()
}

// BeforeShutdown implements lifecycle.BeforeShutdown: saves the state of the active scene.
func (s *SceneService) BeforeShutdown(ctx context.Context) error {
	return s.Exit(ctx)
}

// Exit saves the state of the active scene and deactivates it. Call before shutdown.
func (s *SceneService) Exit(ctx context.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.exit(ctx)
}

func (s *SceneService) load(ctx context.Context, sceneID string) (*scenes.Scene, error) {
	cfg, err := s.configs.GetSceneConfig(ctx, sceneID)
	if err != nil {
		return nil, fmt.Errorf("loading config of scene %q: %w", sceneID, err)
	}

	state, err := s.states.LoadSceneState(ctx, sceneID)
	if errors.Is(err, scenes.ErrStateNotFound) {
		state, err = nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("loading state of scene %q: %w", sceneID, err)
	}

	return scenes.Load(cfg, state)
}

func (s *SceneService) exit(ctx context.Context) error {
	if s.active == nil {
		return nil
	}
	if err := s.states.SaveSceneState(ctx, s.active.State); err != nil {
		return fmt.Errorf("saving state of scene %q: %w", s.active.ID(), err)
	}
	s.active = nil

	return nil
}

// change mutates the active scene and instructs the devices of what changed.
func (s *SceneService) change(ctx context.Context, mutate func(*scenes.Scene) error) (*scenes.Transition, error) {
	if s.active == nil {
		return nil, ErrNoActiveScene
	}
	from := s.active.Ref()
	prev, err := s.active.Instructions()
	if err != nil {
		return nil, err
	}
	if err := mutate(s.active); err != nil {
		return nil, err
	}

	return s.transition(ctx, from, prev, s.active)
}

// transition makes next the active scene and dispatches the changes. Device errors are joined and
// returned; the scene is active regardless.
func (s *SceneService) transition(ctx context.Context, from scenes.Ref, prev scenes.Instructions, next *scenes.Scene) (*scenes.Transition, error) {
	ins, err := next.Instructions()
	if err != nil {
		return nil, err
	}
	room, err := s.rooms.CurrentRoom(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading room: %w", err)
	}

	t := scenes.NewTransition(from, prev, next.Ref(), ins)
	s.active = next

	// TODO: handle t.Released, e.g. blank screens / stop music for targets the new scene no longer uses.
	return t, s.dispatch(ctx, room, t.To, t.Changes)
}

func (s *SceneService) dispatch(ctx context.Context, room *devices.Room, scene scenes.Ref, ins scenes.Instructions) error {
	var err error
	for _, ti := range ins.All() {
		for _, device := range room.DevicesFor(ti.Target, ti.Instruction.Kind()) {
			driver, ok := s.drivers[device.Integration]
			if !ok {
				err = errors.Join(err, fmt.Errorf("device %q (%s): %w", device.ID, device.Integration, ErrNoDriver))

				continue
			}
			if e := driver.Apply(ctx, scene, device, ti.Instruction); e != nil {
				err = errors.Join(err, fmt.Errorf("device %q: %w", device.ID, e))
			}
		}
	}

	return err
}
