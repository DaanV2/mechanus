package application

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
)

var (
	// ErrNoDriver is returned when a device uses an integration that has no registered driver.
	ErrNoDriver = errors.New("no driver registered for integration")
	// ErrNoActiveScene is returned when an operation requires an active scene.
	ErrNoActiveScene = errors.New("no active scene")
	// ErrNoBattleMap is returned when a screen device shows no battle map of the active scene.
	ErrNoBattleMap = errors.New("device shows no battle map of the active scene")
	// ErrDeviceUpdate wraps the errors of devices that could not be told about a change that did happen.
	ErrDeviceUpdate = errors.New("not every device could be updated")
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
	// scene the instruction comes from, e.g. to resolve scene relative asset paths, target the name it was
	// addressed to. Devices are instructed concurrently, one instruction at a time per device, so Apply
	// must be safe for concurrent use and return when ctx is done.
	Apply(ctx context.Context, scene scenes.Ref, target string, device *devices.Device, instruction scenes.Instruction) error
}

// SceneListener is told when the active scene or variant changed, after the devices were instructed.
// A failing listener has no influence on the change, so it handles its own errors.
type SceneListener interface {
	SceneChanged(ctx context.Context, transition *scenes.Transition)
}

// BattleMapDriver applies small battle map changes to a device without re-sending its whole
// instruction, so a moved token or camera doesn't reload the map. Devices of an integration without
// one see these changes with their next instruction.
type BattleMapDriver interface {
	// Integration returns the integration identifier this driver handles, see devices.Integration*.
	Integration() string
	// MoveEntity moves an entity's token, in grid units. Never called for hidden entities.
	MoveEntity(ctx context.Context, scene scenes.Ref, device *devices.Device, entityID string, position vectors.Vector2D) error
	// SetCamera moves the part of the battle map the device shows.
	SetCamera(ctx context.Context, scene scenes.Ref, device *devices.Device, camera scenes.Camera) error
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

	lock       sync.Mutex
	active     *scenes.Scene
	listeners  []SceneListener
	battleMaps map[string]BattleMapDriver
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

// AddListener registers a listener for scene and variant changes.
func (s *SceneService) AddListener(listener SceneListener) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.listeners = append(s.listeners, listener)
}

// AddBattleMapDriver registers the driver for live battle map changes of its integration.
func (s *SceneService) AddBattleMapDriver(driver BattleMapDriver) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.battleMaps == nil {
		s.battleMaps = map[string]BattleMapDriver{}
	}
	s.battleMaps[driver.Integration()] = driver
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
	var from scenes.Ref
	if s.active != nil {
		from = s.active.Ref()
	}
	if err := s.exit(ctx); err != nil {
		return nil, err
	}
	if err := s.states.SaveSceneState(ctx, next.State); err != nil {
		return nil, fmt.Errorf("saving state of scene %q: %w", sceneID, err)
	}

	// No previous instructions: a new scene instructs every device, even where nothing would change.
	return s.transition(ctx, from, scenes.Instructions{}, next)
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

// SetEntityHidden hides an entity of the active scene from the screens, or reveals it.
func (s *SceneService) SetEntityHidden(ctx context.Context, entityID string, hidden bool) (*scenes.Transition, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.change(ctx, func(sc *scenes.Scene) error { return sc.State.SetHidden(entityID, hidden) })
}

// SpawnEntity adds an entity to the active scene during play, at position when given. Without a
// position, and not placed by the battle map, it isn't drawn on battle maps until it is moved.
func (s *SceneService) SpawnEntity(ctx context.Context, entity scenes.Entity, position *vectors.Vector2D) (*scenes.Transition, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Checked first, so a refused spawn leaves no entity behind
	if position != nil {
		if err := scenes.ValidatePosition(*position); err != nil {
			return nil, err
		}
	}

	return s.change(ctx, func(sc *scenes.Scene) error {
		if err := sc.State.Spawn(entity); err != nil {
			return err
		}
		if position == nil {
			return nil
		}

		return sc.State.Move(entity.ID, *position)
	})
}

// RemoveEntity removes an entity from the active scene (defeated, fled, ...).
func (s *SceneService) RemoveEntity(ctx context.Context, entityID string) (*scenes.Transition, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.change(ctx, func(sc *scenes.Scene) error { return sc.State.Remove(entityID) })
}

// MoveEntity moves an entity of the active scene on the battle map, in grid units. The screens
// showing a battle map of the scene get only the moved token, unless the entity had no position in
// the state yet: then those screens are re-instructed once, as the battle map may not have drawn it.
// A hidden entity is moved without telling any screen.
//
// An error wrapping [ErrDeviceUpdate] means the entity was moved but some devices could not be told.
func (s *SceneService) MoveEntity(ctx context.Context, entityID string, position vectors.Vector2D) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.active == nil {
		return ErrNoActiveScene
	}
	if _, moved := s.active.Position(entityID); !moved {
		t, err := s.change(ctx, func(sc *scenes.Scene) error { return sc.State.Move(entityID, position) })

		return deviceUpdateError(t != nil, err)
	}

	if err := s.active.State.Move(entityID, position); err != nil {
		return err
	}
	if s.active.State.Entities[entityID].Hidden {
		return nil
	}
	targets, err := s.active.BattleMapTargets()
	if err != nil {
		return err
	}
	scene := s.active.Ref()

	return s.updateBattleMaps(ctx, targets, func(driver BattleMapDriver, device *devices.Device) error {
		return driver.MoveEntity(ctx, scene, device, entityID, position)
	})
}

// SetCamera moves the camera of the battle maps a screen device shows in the active scene. Every
// device bound to the same targets follows. Returns [ErrNoBattleMap] when the device shows none.
//
// An error wrapping [ErrDeviceUpdate] means the camera was moved but some devices could not be told.
func (s *SceneService) SetCamera(ctx context.Context, deviceID string, camera scenes.Camera) error {
	if err := camera.Validate(); err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	targets, err := s.battleMapTargetsOf(ctx, deviceID)
	if err != nil {
		return err
	}
	for _, target := range targets {
		if err := s.active.State.SetCamera(target, camera); err != nil {
			return err
		}
	}
	scene := s.active.Ref()

	return s.updateBattleMaps(ctx, targets, func(driver BattleMapDriver, device *devices.Device) error {
		return driver.SetCamera(ctx, scene, device, camera)
	})
}

// ShowsBattleMap reports whether a screen device shows a battle map of the active scene.
func (s *SceneService) ShowsBattleMap(ctx context.Context, deviceID string) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, err := s.battleMapTargetsOf(ctx, deviceID)
	switch {
	case errors.Is(err, ErrNoBattleMap), errors.Is(err, ErrNoActiveScene):
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

// ActiveEntities returns the entities present in the active scene, hidden ones included, nil when no
// scene is active.
func (s *SceneService) ActiveEntities() []ActiveEntity {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.active == nil {
		return nil
	}

	entities := s.active.Entities()
	result := make([]ActiveEntity, 0, len(entities))
	for _, e := range entities {
		ae := ActiveEntity{Entity: e}
		if pos, ok := s.active.Position(e.ID); ok {
			ae.Position = &pos
		}
		result = append(result, ae)
	}

	return result
}

// ActiveEntity is an entity present in the active scene.
type ActiveEntity struct {
	scenes.Entity

	Position *vectors.Vector2D // Where it was moved to during play, nil when it stands where the battle map places it
}

// battleMapTargetsOf returns the targets of the active scene that show a battle map on the device.
func (s *SceneService) battleMapTargetsOf(ctx context.Context, deviceID string) ([]string, error) {
	if s.active == nil {
		return nil, ErrNoActiveScene
	}
	targets, err := s.active.BattleMapTargets()
	if err != nil {
		return nil, err
	}
	room, err := s.rooms.CurrentRoom(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading room: %w", err)
	}

	var result []string
	for _, target := range targets {
		for _, device := range room.DevicesFor(target, devices.KindScreen) {
			if device.ID == deviceID {
				result = append(result, target)

				break
			}
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("device %q: %w", deviceID, ErrNoBattleMap)
	}

	return result, nil
}

// updateBattleMaps hands a live change to the battle map driver of every screen device bound to the targets.
func (s *SceneService) updateBattleMaps(ctx context.Context, targets []string, update func(BattleMapDriver, *devices.Device) error) error {
	room, err := s.rooms.CurrentRoom(ctx)
	if err != nil {
		return fmt.Errorf("%w: loading room: %w", ErrDeviceUpdate, err)
	}

	var errs []error
	seen := map[string]bool{}
	for _, target := range targets {
		for _, device := range room.DevicesFor(target, devices.KindScreen) {
			if seen[device.ID] {
				continue
			}
			seen[device.ID] = true

			driver, ok := s.battleMaps[device.Integration]
			if !ok {
				continue
			}
			if err := update(driver, device); err != nil {
				errs = append(errs, fmt.Errorf("device %q: %w", device.ID, err))
			}
		}
	}

	return deviceUpdateError(true, errors.Join(errs...))
}

// deviceUpdateError marks err as a device error when the change itself happened.
func deviceUpdateError(changed bool, err error) error {
	if err == nil || !changed {
		return err
	}

	return fmt.Errorf("%w: %w", ErrDeviceUpdate, err)
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
	err = s.dispatch(ctx, room, t.To, t.Changes)
	if t.From != t.To {
		for _, l := range s.listeners {
			l.SceneChanged(ctx, t)
		}
	}

	return t, err
}

// dispatch instructs every device concurrently, so a slow or unreachable device does not hold up the
// others. A device gets its instructions in order, one at a time.
func (s *SceneService) dispatch(ctx context.Context, room *devices.Room, scene scenes.Ref, ins scenes.Instructions) error {
	queues := deviceQueues(room, ins)
	errs := make([]error, len(queues))

	var wg sync.WaitGroup
	for i, queue := range queues {
		wg.Go(func() { errs[i] = s.applyQueue(ctx, scene, queue) })
	}
	wg.Wait()

	return errors.Join(errs...)
}

func (s *SceneService) applyQueue(ctx context.Context, scene scenes.Ref, queue deviceQueue) error {
	driver, ok := s.drivers[queue.device.Integration]
	if !ok {
		return fmt.Errorf("device %q (%s): %w", queue.device.ID, queue.device.Integration, ErrNoDriver)
	}

	var errs []error
	for _, ti := range queue.instructions {
		if err := driver.Apply(ctx, scene, ti.Target, queue.device, ti.Instruction); err != nil {
			errs = append(errs, fmt.Errorf("device %q: %w", queue.device.ID, err))
		}
	}

	return errors.Join(errs...)
}

// deviceQueue holds the instructions for a single device.
type deviceQueue struct {
	device       *devices.Device
	instructions []scenes.TargetedInstruction
}

// deviceQueues groups instructions by the device they are for, keeping the order of [scenes.Instructions.All].
func deviceQueues(room *devices.Room, ins scenes.Instructions) []deviceQueue {
	var queues []deviceQueue
	index := map[string]int{}
	for _, ti := range ins.All() {
		for _, device := range room.DevicesFor(ti.Target, ti.Instruction.Kind()) {
			i, ok := index[device.ID]
			if !ok {
				i = len(queues)
				index[device.ID] = i
				queues = append(queues, deviceQueue{device: device})
			}
			queues[i].instructions = append(queues[i].instructions, ti)
		}
	}

	return queues
}
