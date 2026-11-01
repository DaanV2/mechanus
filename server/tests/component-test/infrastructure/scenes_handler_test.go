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
	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SceneServiceHandler", func() {
	var (
		handler       *grpc.SceneServiceHandler
		screenManager *screens.ScreenManager
		userContext   func(context.Context) context.Context
	)

	BeforeEach(func(setupCtx SpecContext) {
		// The adapter logic is storage agnostic; the database stands in for files to avoid touching the user's config dir.
		db := util_test.CreateDatabase(setupCtx)
		sceneStorage, err := storage.NamedDBStorage[*scenes.Config](db, "scenes").AppStorage()
		Expect(err).ToNot(HaveOccurred())
		roomStorage, err := storage.NamedDBStorage[*devices.Room](db, "rooms").AppStorage()
		Expect(err).ToNot(HaveOccurred())

		Expect(sceneStorage.Set(setupCtx, &scenes.Config{
			ID:   "cave",
			Name: "Cave",
			Instructions: scenes.Instructions{
				Screens: map[string]scenes.ScreenInstruction{"table": {View: scenes.ViewBattleMap, Asset: "maps/cave.png"}},
			},
			Variants: []scenes.Variant{{
				ID:   "dark",
				Name: "Dark",
				Instructions: scenes.Instructions{
					Screens: map[string]scenes.ScreenInstruction{"table": {View: scenes.ViewBlank}},
				},
			}},
		})).To(Succeed())

		room := devices.NewRoom(scenestore.RoomID, "Living room")
		room.AddDevice(&devices.Device{ID: "tv", Kind: devices.KindScreen, Integration: devices.IntegrationScreen})
		Expect(room.Bind("table", "tv")).To(Succeed())
		Expect(roomStorage.Set(setupCtx, room)).To(Succeed())

		stateStore, err := scenestore.NewDBStateStore(db)
		Expect(err).ToNot(HaveOccurred())
		configStore := scenestore.NewConfigStore(sceneStorage, roomStorage, scenestore.RoomID)
		assetStore := assets.NewStore(fstest.MapFS{"scenes/cave/maps/cave.png": {Data: []byte("png")}})
		screenManager = screens.NewScreenManager()

		handler = grpc.NewSceneServiceHandler(application.NewSceneService(configStore, stateStore, configStore,
			screen.NewDriver(screenManager, assetStore),
		))
		userContext = withUser(util_test.CreateUser())
	})

	screenView := func() *screensv1.View {
		GinkgoHelper()
		h, ok := screenManager.Get("tv")
		Expect(ok).To(BeTrue(), "the tv should have a screen")

		return h.State().GetView()
	}

	It("lists the configured scenes with their variants", func(ctx SpecContext) {
		resp, err := handler.List(userContext(ctx), connect.NewRequest(&scenesv1.ListScenesRequest{}))
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Msg.GetScenes()).To(HaveLen(1))
		Expect(resp.Msg.GetScenes()[0].GetName()).To(Equal("Cave"))
		Expect(resp.Msg.GetScenes()[0].GetVariants()[0].GetId()).To(Equal("dark"))
	})

	It("has no active scene before one is activated", func(ctx SpecContext) {
		resp, err := handler.GetActive(userContext(ctx), connect.NewRequest(&scenesv1.GetActiveSceneRequest{}))
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Msg.GetActive()).To(BeNil())
	})

	It("shows the activated scene's battle map on the bound screen", func(ctx SpecContext) {
		resp, err := handler.Activate(userContext(ctx), connect.NewRequest(&scenesv1.ActivateSceneRequest{SceneId: "cave"}))
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Msg.GetDeviceErrors()).To(BeEmpty())
		Expect(resp.Msg.GetActive().GetScene().GetId()).To(Equal("cave"))

		view := screenView()
		Expect(view.GetKind()).To(Equal(screensv1.ViewKind_VIEW_KIND_BATTLEMAP))
		Expect(view.GetAsset().GetUrl()).To(HavePrefix("/assets/scenes/cave/maps/cave.png?v="))
	})

	It("updates the screen when switching variant", func(ctx SpecContext) {
		_, err := handler.Activate(userContext(ctx), connect.NewRequest(&scenesv1.ActivateSceneRequest{SceneId: "cave"}))
		Expect(err).ToNot(HaveOccurred())

		resp, err := handler.SetVariant(userContext(ctx), connect.NewRequest(&scenesv1.SetVariantRequest{VariantId: "dark"}))
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Msg.GetActive().GetVariantId()).To(Equal("dark"))
		Expect(screenView().GetKind()).To(Equal(screensv1.ViewKind_VIEW_KIND_BLANK))
	})

	DescribeTable("maps failures to connect codes",
		func(ctx SpecContext, call func(context.Context) error, code connect.Code) {
			Expect(connect.CodeOf(call(userContext(ctx)))).To(Equal(code))
		},
		Entry("unknown scene", func(ctx context.Context) error {
			_, err := handler.Activate(ctx, connect.NewRequest(&scenesv1.ActivateSceneRequest{SceneId: "road"}))

			return err
		}, connect.CodeNotFound),
		Entry("invalid scene id", func(ctx context.Context) error {
			_, err := handler.Activate(ctx, connect.NewRequest(&scenesv1.ActivateSceneRequest{SceneId: "../etc"}))

			return err
		}, connect.CodeInvalidArgument),
		Entry("variant without an active scene", func(ctx context.Context) error {
			_, err := handler.SetVariant(ctx, connect.NewRequest(&scenesv1.SetVariantRequest{VariantId: "dark"}))

			return err
		}, connect.CodeFailedPrecondition),
	)

	It("requires an authenticated user", func(ctx SpecContext) {
		_, err := handler.List(ctx, connect.NewRequest(&scenesv1.ListScenesRequest{}))
		Expect(connect.CodeOf(err)).To(Equal(connect.CodeUnauthenticated))
	})
})
