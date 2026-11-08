package infrastructure_test

import (
	"context"
	"testing/fstest"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/infrastructure/assets"
	"github.com/DaanV2/mechanus/server/infrastructure/outputs/screen"
	"github.com/DaanV2/mechanus/server/infrastructure/scenestore"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"
	scenesv1 "github.com/DaanV2/mechanus/server/proto/scenes/v1"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SceneServiceHandler entity controls", func() {
	var (
		handler       *grpc.SceneServiceHandler
		sceneService  *application.SceneService
		screenManager *screens.ScreenManager
		userContext   func(context.Context) context.Context
	)

	BeforeEach(func(setupCtx SpecContext) {
		db := util_test.CreateDatabase(setupCtx)
		sceneStorage, err := storage.NamedDBStorage[*scenes.Config](db, "scenes").AppStorage()
		Expect(err).ToNot(HaveOccurred())
		roomStorage, err := storage.NamedDBStorage[*devices.Room](db, "rooms").AppStorage()
		Expect(err).ToNot(HaveOccurred())

		Expect(sceneStorage.Set(setupCtx, &scenes.Config{
			ID:   "cave",
			Name: "Cave",
			Entities: []scenes.Entity{
				{ID: "hero", Kind: scenes.EntityPlayer, Name: "Hero"},
				{ID: "troll", Kind: scenes.EntityEnemy, Name: "Troll", Hidden: true},
			},
			Instructions: scenes.Instructions{
				Screens: map[string]scenes.ScreenInstruction{"table": {View: scenes.ViewBattleMap, Asset: "maps/cave.png"}},
			},
		})).To(Succeed())

		room := devices.NewRoom(scenestore.RoomID, "Living room")
		room.AddDevice(&devices.Device{ID: "tv", Kind: devices.KindScreen, Integration: devices.IntegrationScreen})
		Expect(room.Bind("table", "tv")).To(Succeed())
		Expect(roomStorage.Set(setupCtx, room)).To(Succeed())

		stateStore, err := scenestore.NewDBStateStore(db)
		Expect(err).ToNot(HaveOccurred())
		configStore := scenestore.NewConfigStore(sceneStorage, roomStorage, scenestore.RoomID)
		screenManager = screens.NewScreenManager()
		driver := screen.NewDriver(screenManager, assets.NewStore(fstest.MapFS{"scenes/cave/maps/cave.png": {Data: []byte("png")}}))

		sceneService = application.NewSceneService(configStore, stateStore, configStore, driver)
		sceneService.AddBattleMapDriver(driver)
		handler = grpc.NewSceneServiceHandler(sceneService, application.NewDeviceMonitor(configStore, 0))
		userContext = withUser(util_test.CreateUser())

		_, err = handler.Activate(userContext(setupCtx), connect.NewRequest(&scenesv1.ActivateSceneRequest{SceneId: "cave"}))
		Expect(err).ToNot(HaveOccurred())
	})

	// screenEntities returns the IDs of the entities the tv draws.
	screenEntities := func() []string {
		GinkgoHelper()
		h, ok := screenManager.Get("tv")
		Expect(ok).To(BeTrue())

		var ids []string
		for _, l := range h.State().GetLayer() {
			for _, e := range l.GetEntity() {
				ids = append(ids, e.GetId())
			}
		}

		return ids
	}

	entity := func(active *scenesv1.ActiveScene, id string) *scenesv1.Entity {
		GinkgoHelper()
		for _, e := range active.GetEntities() {
			if e.GetId() == id {
				return e
			}
		}
		Fail("no entity " + id)

		return nil
	}

	It("lists the entities of the active scene, hidden ones included", func(ctx SpecContext) {
		resp, err := handler.GetActive(userContext(ctx), connect.NewRequest(&scenesv1.GetActiveSceneRequest{}))
		Expect(err).ToNot(HaveOccurred())

		troll := entity(resp.Msg.GetActive(), "troll")
		Expect(troll.GetHidden()).To(BeTrue())
		Expect(troll.GetKind()).To(Equal(scenesv1.EntityKind_ENTITY_KIND_ENEMY))
		Expect(troll.GetPosition()).To(BeNil())
	})

	It("reveals and hides entities", func(ctx SpecContext) {
		resp, err := handler.SetEntityHidden(userContext(ctx), connect.NewRequest(&scenesv1.SetEntityHiddenRequest{EntityId: "troll", Hidden: false}))
		Expect(err).ToNot(HaveOccurred())
		Expect(entity(resp.Msg.GetActive(), "troll").GetHidden()).To(BeFalse())

		_, err = handler.SetEntityHidden(userContext(ctx), connect.NewRequest(&scenesv1.SetEntityHiddenRequest{EntityId: "hero", Hidden: true}))
		Expect(err).ToNot(HaveOccurred())
		Expect(sceneService.ActiveEntities()).To(ContainElement(HaveField("Entity.Hidden", BeTrue())))
	})

	It("spawns entities, drawn on the battle map at their position", func(ctx SpecContext) {
		resp, err := handler.SpawnEntity(userContext(ctx), connect.NewRequest(&scenesv1.SpawnEntityRequest{
			EntityId: "wolf",
			Kind:     scenesv1.EntityKind_ENTITY_KIND_ENEMY,
			Name:     "Wolf",
			Position: &scenesv1.Position{X: 2, Y: 3},
		}))
		Expect(err).ToNot(HaveOccurred())

		wolf := entity(resp.Msg.GetActive(), "wolf")
		Expect(wolf.GetName()).To(Equal("Wolf"))
		Expect(wolf.GetPosition().GetX()).To(Equal(2.0))
		Expect(screenEntities()).To(ConsistOf("wolf"), "the map image places no entities, the spawned one has a position")
	})

	It("removes entities", func(ctx SpecContext) {
		resp, err := handler.RemoveEntity(userContext(ctx), connect.NewRequest(&scenesv1.RemoveEntityRequest{EntityId: "hero"}))
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Msg.GetActive().GetEntities()).To(HaveLen(1))
	})

	It("moves entities, sending the new position to the screens", func(ctx SpecContext) {
		for _, x := range []float64{1, 4} {
			resp, err := handler.MoveEntity(userContext(ctx), connect.NewRequest(&scenesv1.MoveEntityRequest{
				EntityId: "hero",
				Position: &scenesv1.Position{X: x, Y: 1},
			}))
			Expect(err).ToNot(HaveOccurred())
			Expect(entity(resp.Msg.GetActive(), "hero").GetPosition().GetX()).To(Equal(x))
		}

		// The first move draws the hero, which the map image doesn't place; the second only moves its token
		h, _ := screenManager.Get("tv")
		var x []float64
		for _, l := range h.State().GetLayer() {
			for _, e := range l.GetEntity() {
				x = append(x, e.GetPosition().GetX())
			}
		}
		Expect(x).To(Equal([]float64{4}))
	})

	It("resets the active scene to its configuration, also in the database", func(ctx SpecContext) {
		_, err := handler.SpawnEntity(userContext(ctx), connect.NewRequest(&scenesv1.SpawnEntityRequest{
			EntityId: "wolf",
			Kind:     scenesv1.EntityKind_ENTITY_KIND_ENEMY,
			Position: &scenesv1.Position{X: 2, Y: 3},
		}))
		Expect(err).ToNot(HaveOccurred())
		Expect(screenEntities()).To(ConsistOf("wolf"))

		resp, err := handler.Reset(userContext(ctx), connect.NewRequest(&scenesv1.ResetSceneRequest{SceneId: " cave "}))
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Msg.GetDeviceErrors()).To(BeEmpty())
		Expect(resp.Msg.GetActive().GetEntities()).To(HaveLen(2))
		Expect(screenEntities()).To(BeEmpty(), "the tv no longer draws the spawned wolf")

		// Exiting and activating again loads the reset state from the database
		Expect(sceneService.Exit(ctx)).To(Succeed())
		active, err := handler.Activate(userContext(ctx), connect.NewRequest(&scenesv1.ActivateSceneRequest{SceneId: "cave"}))
		Expect(err).ToNot(HaveOccurred())
		Expect(active.Msg.GetActive().GetEntities()).To(HaveLen(2))
	})

	It("resets a scene that isn't active without activating it", func(ctx SpecContext) {
		Expect(sceneService.Exit(ctx)).To(Succeed())

		resp, err := handler.Reset(userContext(ctx), connect.NewRequest(&scenesv1.ResetSceneRequest{SceneId: "cave"}))
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Msg.GetActive()).To(BeNil())
		Expect(sceneService.Active().IsZero()).To(BeTrue())
	})

	DescribeTable("maps failures to connect codes",
		func(ctx SpecContext, call func(context.Context) error, code connect.Code) {
			Expect(connect.CodeOf(call(userContext(ctx)))).To(Equal(code))
		},
		Entry("hiding an unknown entity", func(ctx context.Context) error {
			_, err := handler.SetEntityHidden(ctx, connect.NewRequest(&scenesv1.SetEntityHiddenRequest{EntityId: "ghost"}))

			return err
		}, connect.CodeNotFound),
		Entry("removing an unknown entity", func(ctx context.Context) error {
			_, err := handler.RemoveEntity(ctx, connect.NewRequest(&scenesv1.RemoveEntityRequest{EntityId: "ghost"}))

			return err
		}, connect.CodeNotFound),
		Entry("spawning an existing entity", func(ctx context.Context) error {
			_, err := handler.SpawnEntity(ctx, connect.NewRequest(&scenesv1.SpawnEntityRequest{EntityId: "hero", Kind: scenesv1.EntityKind_ENTITY_KIND_PLAYER}))

			return err
		}, connect.CodeAlreadyExists),
		Entry("spawning without a kind", func(ctx context.Context) error {
			_, err := handler.SpawnEntity(ctx, connect.NewRequest(&scenesv1.SpawnEntityRequest{EntityId: "imp"}))

			return err
		}, connect.CodeInvalidArgument),
		Entry("moving without a position", func(ctx context.Context) error {
			_, err := handler.MoveEntity(ctx, connect.NewRequest(&scenesv1.MoveEntityRequest{EntityId: "hero"}))

			return err
		}, connect.CodeInvalidArgument),
		Entry("moving off the map", func(ctx context.Context) error {
			_, err := handler.MoveEntity(ctx, connect.NewRequest(&scenesv1.MoveEntityRequest{EntityId: "hero", Position: &scenesv1.Position{X: 1e9}}))

			return err
		}, connect.CodeInvalidArgument),
		Entry("resetting an unknown scene", func(ctx context.Context) error {
			_, err := handler.Reset(ctx, connect.NewRequest(&scenesv1.ResetSceneRequest{SceneId: "ghost"}))

			return err
		}, connect.CodeNotFound),
		Entry("resetting without a scene", func(ctx context.Context) error {
			_, err := handler.Reset(ctx, connect.NewRequest(&scenesv1.ResetSceneRequest{}))

			return err
		}, connect.CodeInvalidArgument),
	)

	It("requires an authenticated user", func(ctx SpecContext) {
		_, err := handler.SetEntityHidden(ctx, connect.NewRequest(&scenesv1.SetEntityHiddenRequest{EntityId: "troll"}))
		Expect(connect.CodeOf(err)).To(Equal(connect.CodeUnauthenticated))
		_, err = handler.SpawnEntity(ctx, connect.NewRequest(&scenesv1.SpawnEntityRequest{EntityId: "wolf"}))
		Expect(connect.CodeOf(err)).To(Equal(connect.CodeUnauthenticated))
		_, err = handler.RemoveEntity(ctx, connect.NewRequest(&scenesv1.RemoveEntityRequest{EntityId: "hero"}))
		Expect(connect.CodeOf(err)).To(Equal(connect.CodeUnauthenticated))
		_, err = handler.MoveEntity(ctx, connect.NewRequest(&scenesv1.MoveEntityRequest{EntityId: "hero"}))
		Expect(connect.CodeOf(err)).To(Equal(connect.CodeUnauthenticated))
		_, err = handler.Reset(ctx, connect.NewRequest(&scenesv1.ResetSceneRequest{SceneId: "cave"}))
		Expect(connect.CodeOf(err)).To(Equal(connect.CodeUnauthenticated))
	})
})
