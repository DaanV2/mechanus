package screen_test

import (
	"context"
	"errors"
	"testing/fstest"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/engine/screens/state"
	"github.com/DaanV2/mechanus/server/infrastructure/assets"
	"github.com/DaanV2/mechanus/server/infrastructure/outputs/screen"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type oneScene struct{ cfg *scenes.Config }

func (s oneScene) GetSceneConfig(_ context.Context, id string) (*scenes.Config, error) {
	if id != s.cfg.ID {
		return nil, scenes.ErrSceneNotFound
	}

	return s.cfg, nil
}

func (s oneScene) ListSceneConfigs(context.Context) ([]*scenes.Config, error) {
	return []*scenes.Config{s.cfg}, nil
}

type noStates struct{}

func (noStates) LoadSceneState(context.Context, string) (*scenes.State, error) {
	return nil, scenes.ErrStateNotFound
}

func (noStates) SaveSceneState(context.Context, *scenes.State) error { return nil }

type fixedRoom struct{ room *devices.Room }

func (r fixedRoom) CurrentRoom(context.Context) (*devices.Room, error) { return r.room, nil }

func refusedCode(err error) screensv1.ActionErrorCode {
	GinkgoHelper()
	var refused *screens.RefusedError
	Expect(errors.As(err, &refused)).To(BeTrue(), "a refusal, got %v", err)

	return refused.Code
}

var _ = Describe("Controller", func() {
	var (
		manager    *screens.ScreenManager
		controller *screen.Controller
		user       *screens.ConnectionInfo
		device     *screens.ConnectionInfo
	)

	BeforeEach(func(ctx SpecContext) {
		room := devices.NewRoom("room", "Room")
		room.AddDevice(&devices.Device{ID: "tv", Kind: devices.KindScreen, Integration: devices.IntegrationScreen})
		room.AddDevice(&devices.Device{ID: "wall", Kind: devices.KindScreen, Integration: devices.IntegrationScreen})
		Expect(room.Bind("table", "tv")).To(Succeed())
		Expect(room.Bind("wall", "wall")).To(Succeed())

		manager = screens.NewScreenManager()
		driver := screen.NewDriver(manager, assets.NewStore(fstest.MapFS{"scenes/tavern/map.png": {Data: []byte("map")}}))
		svc := application.NewSceneService(oneScene{&scenes.Config{
			ID:       "tavern",
			Entities: []scenes.Entity{{ID: "hero", Kind: scenes.EntityPlayer, Name: "Hero"}},
			Instructions: scenes.Instructions{Screens: map[string]scenes.ScreenInstruction{
				"table": {View: scenes.ViewBattleMap, Asset: "map.png"},
				"wall":  {View: scenes.ViewBlank},
			}},
		}}, noStates{}, fixedRoom{room}, driver)
		svc.AddBattleMapDriver(driver)
		_, err := svc.Activate(ctx, "tavern", "")
		Expect(err).ToNot(HaveOccurred())

		controller = screen.NewController(svc)
		user = &screens.ConnectionInfo{ID: "user-1", Type: devices.DeviceTypeUser, Roles: []string{roles.User.String()}}
		device = &screens.ConnectionInfo{ID: "tv", DeviceID: "tv", Type: devices.DeviceTypeDevice, Roles: []string{roles.Device.String()}}
	})

	Describe("MoveEntity", func() {
		It("moves the token for logged in users", func(ctx SpecContext) {
			Expect(controller.MoveEntity(ctx, user, "tv", "hero", vectors.Vector2D{X: 2, Y: 3})).To(Succeed())

			handler, _ := manager.Get("tv")
			Expect(handler.State().GetLayer()).To(ContainElement(Satisfy(func(l *screensv1.Layer) bool {
				return len(l.GetEntity()) == 1 && l.GetEntity()[0].GetPosition().GetX() == 2
			})), "drawn at its new position")
		})

		It("refuses devices and users without the user role", func(ctx SpecContext) {
			err := controller.MoveEntity(ctx, device, "tv", "hero", vectors.Vector2D{X: 2, Y: 3})
			Expect(refusedCode(err)).To(Equal(screensv1.ActionErrorCode_ACTION_ERROR_CODE_PERMISSION_DENIED))

			viewer := &screens.ConnectionInfo{ID: "user-2", Type: devices.DeviceTypeUser, Roles: []string{roles.Viewer.String()}}
			err = controller.MoveEntity(ctx, viewer, "tv", "hero", vectors.Vector2D{X: 2, Y: 3})
			Expect(refusedCode(err)).To(Equal(screensv1.ActionErrorCode_ACTION_ERROR_CODE_PERMISSION_DENIED))
		})

		It("refuses screens that show no battle map of the scene", func(ctx SpecContext) {
			err := controller.MoveEntity(ctx, user, "wall", "hero", vectors.Vector2D{X: 2, Y: 3})
			Expect(refusedCode(err)).To(Equal(screensv1.ActionErrorCode_ACTION_ERROR_CODE_FAILED_PRECONDITION))
		})

		It("refuses unknown entities and invalid positions", func(ctx SpecContext) {
			err := controller.MoveEntity(ctx, user, "tv", "ghost", vectors.Vector2D{})
			Expect(refusedCode(err)).To(Equal(screensv1.ActionErrorCode_ACTION_ERROR_CODE_NOT_FOUND))

			err = controller.MoveEntity(ctx, user, "tv", "hero", vectors.Vector2D{X: -scenes.MaxCoordinate * 2})
			Expect(refusedCode(err)).To(Equal(screensv1.ActionErrorCode_ACTION_ERROR_CODE_INVALID_ARGUMENT))
		})
	})

	Describe("SetCamera", func() {
		It("lets users and the screen's own device pan and zoom", func(ctx SpecContext) {
			Expect(controller.SetCamera(ctx, user, "tv", state.Camera{Zoom: 2})).To(Succeed())
			Expect(controller.SetCamera(ctx, device, "tv", state.Camera{Center: vectors.Vector2D{X: 1, Y: 1}, Zoom: 3})).To(Succeed())

			handler, _ := manager.Get("tv")
			Expect(handler.State().GetCamera().GetZoom()).To(Equal(3.0))
		})

		It("refuses invalid cameras and screens without a battle map", func(ctx SpecContext) {
			err := controller.SetCamera(ctx, user, "tv", state.Camera{Zoom: 0})
			Expect(refusedCode(err)).To(Equal(screensv1.ActionErrorCode_ACTION_ERROR_CODE_INVALID_ARGUMENT))

			err = controller.SetCamera(ctx, user, "wall", state.Camera{Zoom: 1})
			Expect(refusedCode(err)).To(Equal(screensv1.ActionErrorCode_ACTION_ERROR_CODE_FAILED_PRECONDITION))
		})
	})
})
