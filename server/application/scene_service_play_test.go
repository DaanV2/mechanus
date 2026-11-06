package application_test

import (
	"context"
	"errors"
	"sync"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type liveUpdate struct {
	device   string
	entity   string
	position vectors.Vector2D
	camera   *scenes.Camera
}

type recordingBattleMaps struct {
	lock    sync.Mutex
	updates []liveUpdate
	err     error
}

func (d *recordingBattleMaps) Integration() string { return devices.IntegrationScreen }

func (d *recordingBattleMaps) MoveEntity(_ context.Context, _ scenes.Ref, device *devices.Device, entityID string, position vectors.Vector2D) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.updates = append(d.updates, liveUpdate{device: device.ID, entity: entityID, position: position})

	return d.err
}

func (d *recordingBattleMaps) SetCamera(_ context.Context, _ scenes.Ref, device *devices.Device, camera scenes.Camera) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.updates = append(d.updates, liveUpdate{device: device.ID, camera: &camera})

	return d.err
}

var _ = Describe("SceneService during play", func() {
	var (
		ctx    context.Context
		states *stateStore
		screen *recordingDriver
		live   *recordingBattleMaps
		svc    *application.SceneService
	)

	BeforeEach(func() {
		ctx = context.Background()
		room := devices.NewRoom("room", "Living room")
		for _, id := range []string{"tv", "tablet", "wall-tv"} {
			room.AddDevice(&devices.Device{ID: id, Kind: devices.KindScreen, Integration: devices.IntegrationScreen})
		}
		Expect(room.Bind("table", "tv")).To(Succeed())
		Expect(room.Bind("table", "tablet")).To(Succeed())
		Expect(room.Bind("wall", "wall-tv")).To(Succeed())

		configs := configStore{
			"cave": {
				ID: "cave",
				Entities: []scenes.Entity{
					{ID: "troll", Kind: scenes.EntityEnemy, Hidden: true},
					{ID: "hero", Kind: scenes.EntityPlayer},
				},
				Instructions: scenes.Instructions{
					Screens: map[string]scenes.ScreenInstruction{
						"table": {View: scenes.ViewBattleMap, Asset: "cave-map"},
						"wall":  {View: scenes.ViewScenery, Asset: "cave.png"},
					},
				},
			},
			"road": {ID: "road"},
		}

		states = &stateStore{saved: map[string]scenes.State{}}
		screen = &recordingDriver{integration: devices.IntegrationScreen}
		live = &recordingBattleMaps{}
		svc = application.NewSceneService(configs, states, staticRoom{room}, screen)
		svc.AddBattleMapDriver(live)

		_, err := svc.Activate(ctx, "cave", "")
		Expect(err).ToNot(HaveOccurred())
		screen.applied = nil
	})

	battleMapEntities := func() []string {
		GinkgoHelper()
		Expect(screen.applied).ToNot(BeEmpty())
		ins, ok := screen.applied[len(screen.applied)-1].instruction.(scenes.ScreenInstruction)
		Expect(ok).To(BeTrue())

		ids := make([]string, 0, len(ins.Entities))
		for _, e := range ins.Entities {
			ids = append(ids, e.ID)
		}

		return ids
	}

	Describe("GM controls", func() {
		It("reveals and hides entities, re-instructing only the battle map screens", func() {
			_, err := svc.SetEntityHidden(ctx, "troll", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(screen.applied).To(HaveLen(2), "tv and tablet, not the wall")
			Expect(battleMapEntities()).To(ConsistOf("troll", "hero"))

			_, err = svc.SetEntityHidden(ctx, "troll", true)
			Expect(err).ToNot(HaveOccurred())
			Expect(battleMapEntities()).To(ConsistOf("hero"), "hidden entities are not sent to screens")
		})

		It("spawns entities, at a position when given", func() {
			_, err := svc.SpawnEntity(ctx, scenes.Entity{ID: "wolf", Kind: scenes.EntityEnemy, Name: "Wolf"}, &vectors.Vector2D{X: 2, Y: 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(battleMapEntities()).To(ConsistOf("hero", "wolf"))

			ins, _ := screen.applied[len(screen.applied)-1].instruction.(scenes.ScreenInstruction)
			Expect(ins.Positions).To(HaveKeyWithValue("wolf", vectors.Vector2D{X: 2, Y: 3}))
			Expect(svc.ActiveEntities()).To(ContainElement(application.ActiveEntity{
				Entity:   scenes.Entity{ID: "wolf", Kind: scenes.EntityEnemy, Name: "Wolf"},
				Position: &vectors.Vector2D{X: 2, Y: 3},
			}))
		})

		It("spawns nothing when the entity or its position is invalid", func() {
			_, err := svc.SpawnEntity(ctx, scenes.Entity{ID: "wolf", Kind: scenes.EntityEnemy}, &vectors.Vector2D{X: scenes.MaxCoordinate * 2})
			Expect(err).To(MatchError(scenes.ErrInvalidPosition))
			_, err = svc.SpawnEntity(ctx, scenes.Entity{ID: "hero", Kind: scenes.EntityPlayer}, nil)
			Expect(err).To(MatchError(scenes.ErrEntityExists))
			_, err = svc.SpawnEntity(ctx, scenes.Entity{ID: "imp"}, nil)
			Expect(err).To(MatchError(scenes.ErrInvalidEntity))

			Expect(svc.ActiveEntities()).To(HaveLen(2))
			Expect(screen.applied).To(BeEmpty())
		})

		It("removes entities", func() {
			_, err := svc.RemoveEntity(ctx, "hero")
			Expect(err).ToNot(HaveOccurred())
			Expect(battleMapEntities()).To(BeEmpty())
			Expect(svc.ActiveEntities()).To(ConsistOf(application.ActiveEntity{
				Entity: scenes.Entity{ID: "troll", Kind: scenes.EntityEnemy, Hidden: true},
			}), "the GM still sees hidden entities")
		})

		It("refuses unknown entities and needs an active scene", func() {
			_, err := svc.SetEntityHidden(ctx, "ghost", false)
			Expect(err).To(MatchError(scenes.ErrEntityNotFound))
			_, err = svc.RemoveEntity(ctx, "ghost")
			Expect(err).To(MatchError(scenes.ErrEntityNotFound))

			Expect(svc.Exit(ctx)).To(Succeed())
			_, err = svc.SetEntityHidden(ctx, "troll", false)
			Expect(err).To(MatchError(application.ErrNoActiveScene))
			Expect(svc.MoveEntity(ctx, "hero", vectors.Vector2D{})).To(MatchError(application.ErrNoActiveScene))
			Expect(svc.ActiveEntities()).To(BeNil())
		})
	})

	Describe("MoveEntity", func() {
		It("re-instructs the battle map screens on the first move, as the map may not draw the entity", func() {
			Expect(svc.MoveEntity(ctx, "hero", vectors.Vector2D{X: 1, Y: 1})).To(Succeed())
			Expect(screen.applied).To(HaveLen(2))
			Expect(live.updates).To(BeEmpty())
		})

		It("sends later moves to every battle map screen without re-instructing them", func() {
			Expect(svc.MoveEntity(ctx, "hero", vectors.Vector2D{X: 1, Y: 1})).To(Succeed())
			screen.applied = nil

			Expect(svc.MoveEntity(ctx, "hero", vectors.Vector2D{X: 2, Y: 5})).To(Succeed())
			Expect(screen.applied).To(BeEmpty())
			Expect(live.updates).To(ConsistOf(
				liveUpdate{device: "tv", entity: "hero", position: vectors.Vector2D{X: 2, Y: 5}},
				liveUpdate{device: "tablet", entity: "hero", position: vectors.Vector2D{X: 2, Y: 5}},
			))
		})

		It("moves hidden entities without telling any screen", func() {
			Expect(svc.MoveEntity(ctx, "troll", vectors.Vector2D{X: 1, Y: 1})).To(Succeed())
			Expect(svc.MoveEntity(ctx, "troll", vectors.Vector2D{X: 3, Y: 3})).To(Succeed())
			Expect(live.updates).To(BeEmpty())

			_, err := svc.SetEntityHidden(ctx, "troll", false)
			Expect(err).ToNot(HaveOccurred())
			ins, _ := screen.applied[len(screen.applied)-1].instruction.(scenes.ScreenInstruction)
			Expect(ins.Positions).To(HaveKeyWithValue("troll", vectors.Vector2D{X: 3, Y: 3}), "revealed where it was moved to")
		})

		It("saves the position with the state on exit", func() {
			Expect(svc.MoveEntity(ctx, "hero", vectors.Vector2D{X: 4, Y: 2})).To(Succeed())
			_, err := svc.Activate(ctx, "road", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(states.saved["cave"].Entities["hero"].Position).To(Equal(&vectors.Vector2D{X: 4, Y: 2}))
		})

		It("refuses invalid moves", func() {
			Expect(svc.MoveEntity(ctx, "ghost", vectors.Vector2D{})).To(MatchError(scenes.ErrEntityNotFound))
			Expect(svc.MoveEntity(ctx, "hero", vectors.Vector2D{X: scenes.MaxCoordinate * 2})).To(MatchError(scenes.ErrInvalidPosition))
		})

		It("reports devices that could not be told, the entity moved regardless", func() {
			Expect(svc.MoveEntity(ctx, "hero", vectors.Vector2D{X: 1, Y: 1})).To(Succeed())
			live.err = errors.New("unreachable")

			Expect(svc.MoveEntity(ctx, "hero", vectors.Vector2D{X: 2, Y: 2})).To(MatchError(application.ErrDeviceUpdate))
			Expect(svc.ActiveEntities()).To(ContainElement(HaveField("Position", Equal(&vectors.Vector2D{X: 2, Y: 2}))))
		})
	})

	Describe("SetCamera", func() {
		camera := scenes.Camera{Center: vectors.Vector2D{X: 3, Y: 4}, Zoom: 2}

		It("moves the camera of every screen bound to the same battle map target", func() {
			Expect(svc.SetCamera(ctx, "tablet", camera)).To(Succeed())
			Expect(live.updates).To(ConsistOf(
				liveUpdate{device: "tv", camera: &camera},
				liveUpdate{device: "tablet", camera: &camera},
			))
			Expect(screen.applied).To(BeEmpty(), "the map is not re-sent")
		})

		It("keeps the camera when the battle map is re-instructed", func() {
			Expect(svc.SetCamera(ctx, "tv", camera)).To(Succeed())
			_, err := svc.SetEntityHidden(ctx, "troll", false)
			Expect(err).ToNot(HaveOccurred())

			ins, _ := screen.applied[len(screen.applied)-1].instruction.(scenes.ScreenInstruction)
			Expect(ins.Camera).To(Equal(&camera))
		})

		It("refuses screens without a battle map, and invalid cameras", func() {
			Expect(svc.SetCamera(ctx, "wall-tv", camera)).To(MatchError(application.ErrNoBattleMap))
			Expect(svc.SetCamera(ctx, "unknown", camera)).To(MatchError(application.ErrNoBattleMap))
			Expect(svc.SetCamera(ctx, "tv", scenes.Camera{Zoom: 100})).To(MatchError(scenes.ErrInvalidCamera))
			Expect(live.updates).To(BeEmpty())
		})

		It("tells which screens show a battle map", func() {
			Expect(svc.ShowsBattleMap(ctx, "tv")).To(BeTrue())
			Expect(svc.ShowsBattleMap(ctx, "wall-tv")).To(BeFalse())
			Expect(svc.Exit(ctx)).To(Succeed())
			Expect(svc.ShowsBattleMap(ctx, "tv")).To(BeFalse())
		})
	})
})
