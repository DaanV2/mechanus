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
)

// SceneRepository loads scenes. Implemented in infrastructure (not yet).
type SceneRepository interface {
	GetScene(ctx context.Context, sceneID string) (*scenes.Scene, error)
}

// RoomProvider returns the room setup scenes are played in. Implemented in infrastructure (not yet).
type RoomProvider interface {
	CurrentRoom(ctx context.Context) (*devices.Room, error)
}

// OutputDriver delivers instructions to the devices of one integration (hue, spotify, mqtt, the
// built-in screen client, ...). Implementations live in infrastructure.
type OutputDriver interface {
	// Integration returns the integration identifier this driver handles, see devices.Integration*.
	Integration() string
	// Apply instructs a single device. The instruction kind always matches the device kind.
	Apply(ctx context.Context, device *devices.Device, instruction scenes.Instruction) error
}

// SceneService activates scenes and variants and instructs the devices in the room.
type SceneService struct {
	scenes  SceneRepository
	rooms   RoomProvider
	drivers map[string]OutputDriver

	lock    sync.Mutex
	current *scenes.Scene
	state   scenes.State
}

// NewSceneService creates a new SceneService with the given drivers.
func NewSceneService(repo SceneRepository, rooms RoomProvider, drivers ...OutputDriver) *SceneService {
	s := &SceneService{
		scenes:  repo,
		rooms:   rooms,
		drivers: make(map[string]OutputDriver, len(drivers)),
	}
	for _, d := range drivers {
		s.drivers[d.Integration()] = d
	}

	return s
}

// Current returns the active scene state.
func (s *SceneService) Current() scenes.State {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.state
}

// Activate switches to the given scene and variant (empty for the scene default) and sends the
// changed instructions to every bound device. Activating a variant of the active scene is a variant
// transition. The state is updated even if some devices failed; their errors are returned joined.
func (s *SceneService) Activate(ctx context.Context, sceneID, variantID string) (*scenes.Transition, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	next, err := s.scenes.GetScene(ctx, sceneID)
	if err != nil {
		return nil, fmt.Errorf("loading scene %q: %w", sceneID, err)
	}
	room, err := s.rooms.CurrentRoom(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading room: %w", err)
	}

	t, err := scenes.Plan(s.current, s.state.VariantID, next, variantID)
	if err != nil {
		return nil, fmt.Errorf("planning transition: %w", err)
	}

	s.current = next
	s.state = t.To

	// TODO: handle t.Released, e.g. blank screens / stop music for targets the new scene no longer uses.
	return t, s.dispatch(ctx, room, t.Changes)
}

func (s *SceneService) dispatch(ctx context.Context, room *devices.Room, ins scenes.Instructions) error {
	var err error
	for _, ti := range ins.All() {
		for _, device := range room.DevicesFor(ti.Target, ti.Instruction.Kind()) {
			driver, ok := s.drivers[device.Integration]
			if !ok {
				err = errors.Join(err, fmt.Errorf("device %q (%s): %w", device.ID, device.Integration, ErrNoDriver))

				continue
			}
			if e := driver.Apply(ctx, device, ti.Instruction); e != nil {
				err = errors.Join(err, fmt.Errorf("device %q: %w", device.ID, e))
			}
		}
	}

	return err
}
