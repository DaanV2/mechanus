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
		configStore   *scenestore.ConfigStore
		configDir     string
	)

	BeforeEach(func(setupCtx SpecContext) {
		db := util_test.CreateDatabase(setupCtx)
		cave := &scenes.Config{
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
		}

		room := devices.NewRoom(scenestore.RoomID, "Living room")
		room.AddDevice(&devices.Device{ID: "tv", Kind: devices.KindScreen, Integration: devices.IntegrationScreen})
		Expect(room.Bind("table", "tv")).To(Succeed())

		stateStore, err := scenestore.NewDBStateStore(db)
		Expect(err).ToNot(HaveOccurred())
		configStore, configDir = util_test.CreateConfigStore(setupCtx, room, cave)
		assetStore := assets.NewStore(fstest.MapFS{
			"scenes/cave/maps/cave.png":  {Data: []byte("png")},
			"scenes/cave/maps/cave2.png": {Data: []byte("png2")},
		})
		screenManager = screens.NewScreenManager()

		monitor := application.NewDeviceMonitor(configStore, 0)
		sceneService := application.NewSceneService(configStore, stateStore, configStore,
			monitor.Track(screen.NewDriver(screenManager, assetStore)),
		)
		handler = grpc.NewSceneServiceHandler(sceneService, monitor, configStore)
		// As wired in components: reloaded config files re-instruct the devices
		configStore.AddListener(func(ctx context.Context, change scenestore.ConfigChange) {
			var err error
			if change.Room {
				_, err = sceneService.RoomConfigChanged(ctx)
			} else {
				_, err = sceneService.SceneConfigChanged(ctx, change.SceneID)
			}
			Expect(err).ToNot(HaveOccurred())
		})
		userContext = withUser(util_test.CreateUser())
	})

	screenView := func() *screensv1.View {
		GinkgoHelper()
		h, ok := screenManager.Get("tv")
		Expect(ok).To(BeTrue(), "the tv should have a screen")

		return h.State().GetView()
	}

	Describe("reloaded config files", func() {
		BeforeEach(func(ctx SpecContext) {
			_, err := handler.Activate(userContext(ctx), connect.NewRequest(&scenesv1.ActivateSceneRequest{SceneId: "cave"}))
			Expect(err).ToNot(HaveOccurred())
		})

		It("re-instructs the screen when the active scene's file changed", func(ctx SpecContext) {
			util_test.WriteRawFile(configDir, "scenes/cave.yaml", `
version: 1
name: Cave
instructions:
  screens:
    table: { view: battlemap, asset: maps/cave2.png }
`)
			configStore.Reload(ctx, "scenes/cave.yaml")

			Expect(screenView().GetAsset().GetUrl()).To(HavePrefix("/assets/scenes/cave/maps/cave2.png?v="))
		})

		It("keeps the screen as it is when the edited file is invalid, and reports it in the list", func(ctx SpecContext) {
			util_test.WriteRawFile(configDir, "scenes/cave.yaml", "version: 1\ninstructions: [")
			configStore.Reload(ctx, "scenes/cave.yaml")

			Expect(screenView().GetAsset().GetUrl()).To(HavePrefix("/assets/scenes/cave/maps/cave.png?v="))
			resp, err := handler.List(userContext(ctx), connect.NewRequest(&scenesv1.ListScenesRequest{}))
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Msg.GetScenes()).To(HaveLen(1), "the last good version is still listed")
			Expect(resp.Msg.GetConfigErrors()).To(HaveLen(1))
			Expect(resp.Msg.GetConfigErrors()[0].GetFile()).To(Equal("scenes/cave.yaml"))
			Expect(resp.Msg.GetConfigErrors()[0].GetMessage()).To(ContainSubstring("YAML"))
		})

		It("falls back to the default variant when the active variant was removed", func(ctx SpecContext) {
			_, err := handler.SetVariant(userContext(ctx), connect.NewRequest(&scenesv1.SetVariantRequest{VariantId: "dark"}))
			Expect(err).ToNot(HaveOccurred())
			Expect(screenView().GetKind()).To(Equal(screensv1.ViewKind_VIEW_KIND_BLANK))

			util_test.WriteRawFile(configDir, "scenes/cave.yaml", `
version: 1
name: Cave
instructions:
  screens:
    table: { view: battlemap, asset: maps/cave.png }
`)
			configStore.Reload(ctx, "scenes/cave.yaml")

			Expect(screenView().GetKind()).To(Equal(screensv1.ViewKind_VIEW_KIND_BATTLEMAP))
			active, err := handler.GetActive(userContext(ctx), connect.NewRequest(&scenesv1.GetActiveSceneRequest{}))
			Expect(err).ToNot(HaveOccurred())
			Expect(active.Msg.GetActive().GetVariantId()).To(BeEmpty())
		})

		It("instructs a screen that the reloaded room newly binds", func(ctx SpecContext) {
			room := devices.NewRoom(scenestore.RoomID, "Living room")
			room.AddDevice(&devices.Device{ID: "tv", Kind: devices.KindScreen, Integration: devices.IntegrationScreen})
			room.AddDevice(&devices.Device{ID: "tv2", Kind: devices.KindScreen, Integration: devices.IntegrationScreen})
			Expect(room.Bind("table", "tv")).To(Succeed())
			Expect(room.Bind("table", "tv2")).To(Succeed())
			util_test.WriteConfigFile(configDir, "rooms/room.json", room)
			configStore.Reload(ctx, "rooms/room.json")

			h, ok := screenManager.Get("tv2")
			Expect(ok).To(BeTrue(), "the new tv should have been instructed")
			Expect(h.State().GetView().GetKind()).To(Equal(screensv1.ViewKind_VIEW_KIND_BATTLEMAP))
		})
	})

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

	It("reports the status of the room's devices", func(ctx SpecContext) {
		_, err := handler.Activate(userContext(ctx), connect.NewRequest(&scenesv1.ActivateSceneRequest{SceneId: "cave"}))
		Expect(err).ToNot(HaveOccurred())

		resp, err := handler.ListDeviceStatuses(userContext(ctx), connect.NewRequest(&scenesv1.ListDeviceStatusesRequest{}))
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Msg.GetDevices()).To(HaveLen(1))

		tv := resp.Msg.GetDevices()[0]
		Expect(tv.GetDeviceId()).To(Equal("tv"))
		Expect(tv.GetKind()).To(Equal(string(devices.KindScreen)))
		Expect(tv.GetIntegration()).To(Equal(devices.IntegrationScreen))
		Expect(tv.GetTargets()).To(Equal([]string{"table"}))
		Expect(tv.GetState()).To(Equal(scenesv1.DeviceState_DEVICE_STATE_OFFLINE), "instructed, but its screen isn't open")
		Expect(tv.GetError()).To(Equal(screen.ErrNotConnected.Error()))
		Expect(tv.GetLastApply()).ToNot(BeNil())
		Expect(tv.GetLastApplyError()).To(BeEmpty())
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

		_, err = handler.ListDeviceStatuses(ctx, connect.NewRequest(&scenesv1.ListDeviceStatusesRequest{}))
		Expect(connect.CodeOf(err)).To(Equal(connect.CodeUnauthenticated))
	})
})
