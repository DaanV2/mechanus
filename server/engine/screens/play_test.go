package screens_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"

	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/engine/screens/state"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
	"github.com/coder/websocket"
)

type move struct {
	conn     *screens.ConnectionInfo
	screenID string
	entityID string
	position vectors.Vector2D
}

type fakeController struct {
	lock    sync.Mutex
	moves   []move
	cameras []state.Camera
	err     error
}

func (c *fakeController) MoveEntity(_ context.Context, conn *screens.ConnectionInfo, screenID, entityID string, position vectors.Vector2D) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.moves = append(c.moves, move{conn, screenID, entityID, position})

	return c.err
}

func (c *fakeController) SetCamera(_ context.Context, _ *screens.ConnectionInfo, _ string, camera state.Camera) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cameras = append(c.cameras, camera)

	return c.err
}

// screenClient is a WebSocket client connected to a screen of the manager.
type screenClient struct {
	conn *websocket.Conn
}

func connectScreen(ctx context.Context, manager *screens.ScreenManager, screenID string, info *screens.ConnectionInfo) *screenClient {
	GinkgoHelper()
	connected := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer GinkgoRecover()

		conn, err := websocket.Accept(w, r, nil)
		Expect(err).ToNot(HaveOccurred())

		sc := screens.NewScreenConn("conn-"+screenID, screenID, r.Context(), conn, info)
		manager.GetOrCreate(screenID).AddListener(r.Context(), sc)
		close(connected)
		<-sc.Context().Done()
	}))
	DeferCleanup(server.Close)

	client, resp, err := websocket.Dial(ctx, server.URL, nil) //nolint:bodyclose // closed in the DeferCleanup below
	Expect(err).ToNot(HaveOccurred())
	DeferCleanup(func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	})
	DeferCleanup(func() { _ = client.CloseNow() })
	Eventually(connected).Should(BeClosed())

	return &screenClient{conn: client}
}

func (c *screenClient) send(ctx context.Context, msgs ...*screensv1.ClientMessage) {
	GinkgoHelper()
	data, err := proto.Marshal(&screensv1.ClientMessages{Action: msgs})
	Expect(err).ToNot(HaveOccurred())
	Expect(c.conn.Write(ctx, websocket.MessageBinary, data)).To(Succeed())
}

func (c *screenClient) read(ctx context.Context) *screensv1.ServerMessage {
	GinkgoHelper()
	readCtx, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()

	_, data, err := c.conn.Read(readCtx)
	Expect(err).ToNot(HaveOccurred())
	var msgs screensv1.ServerMessages
	Expect(proto.Unmarshal(data, &msgs)).To(Succeed())
	Expect(msgs.GetAction()).To(HaveLen(1))

	return msgs.GetAction()[0]
}

func moveMessage(id, entityID string, x, y float64) *screensv1.ClientMessage {
	return &screensv1.ClientMessage{
		Id: &id,
		Action: &screensv1.ClientMessage_MoveEntity{MoveEntity: &screensv1.MoveEntity{
			EntityId: entityID,
			Position: &screensv1.Vector2{X: x, Y: y},
		}},
	}
}

func battleMapState() *state.ScreenState {
	st := state.NewScreenState()
	st.View.Kind = scenes.ViewBattleMap
	st.Layers = []*state.Layer{
		{ID: "floor", Kind: scenes.MapLayerImage},
		{ID: "tokens", Kind: scenes.MapLayerEntities, Entities: map[string]*state.Entity{
			"hero": {ID: "hero", Name: "Hero", Kind: scenes.EntityPlayer, Position: vectors.Vector2D{X: 1, Y: 1}, Size: vectors.Vector2D{X: 1, Y: 1}},
		}},
	}

	return st
}

var _ = Describe("Play actions", func() {
	var (
		manager    *screens.ScreenManager
		controller *fakeController
		user       *screens.ConnectionInfo
	)

	BeforeEach(func() {
		manager = screens.NewScreenManager()
		controller = &fakeController{}
		manager.SetController(controller)
		user = &screens.ConnectionInfo{ID: "user-1", Type: devices.DeviceTypeUser}
	})

	It("hands token moves to the controller, with the connection and screen", func(ctx SpecContext) {
		client := connectScreen(ctx, manager, "tv", user)
		client.send(ctx, moveMessage("1", "hero", 3, 4))

		Eventually(func() []move {
			controller.lock.Lock()
			defer controller.lock.Unlock()

			return controller.moves
		}).Should(ConsistOf(move{user, "tv", "hero", vectors.Vector2D{X: 3, Y: 4}}))
	})

	It("hands camera moves to the controller", func(ctx SpecContext) {
		client := connectScreen(ctx, manager, "tv", user)
		client.send(ctx, &screensv1.ClientMessage{Action: &screensv1.ClientMessage_SetCamera{SetCamera: &screensv1.SetCamera{
			Camera: &screensv1.Camera{Center: &screensv1.Vector2{X: 2, Y: 3}, Zoom: 1.5},
		}}})

		Eventually(func() []state.Camera {
			controller.lock.Lock()
			defer controller.lock.Unlock()

			return controller.cameras
		}).Should(ConsistOf(state.Camera{Center: vectors.Vector2D{X: 2, Y: 3}, Zoom: 1.5}))
	})

	It("tells the client why an action was refused, keeping the connection open", func(ctx SpecContext) {
		controller.err = screens.Refuse(screensv1.ActionErrorCode_ACTION_ERROR_CODE_PERMISSION_DENIED, errors.New("not yours"))
		client := connectScreen(ctx, manager, "tv", user)

		client.send(ctx, moveMessage("42", "hero", 3, 4))
		msg := client.read(ctx)
		Expect(msg.GetId()).To(Equal("42"))
		Expect(msg.GetActionError().GetCode()).To(Equal(screensv1.ActionErrorCode_ACTION_ERROR_CODE_PERMISSION_DENIED))
		Expect(msg.GetActionError().GetMessage()).To(Equal("not yours"))

		controller.err = errors.New("database on fire")
		client.send(ctx, moveMessage("43", "hero", 3, 4))
		msg = client.read(ctx)
		Expect(msg.GetActionError().GetCode()).To(Equal(screensv1.ActionErrorCode_ACTION_ERROR_CODE_INTERNAL))
		Expect(msg.GetActionError().GetMessage()).To(Equal("internal error"), "internal errors are not shown to clients")
	})

	It("refuses play actions without a controller", func(ctx SpecContext) {
		client := connectScreen(ctx, screens.NewScreenManager(), "tv", user)
		client.send(ctx, moveMessage("1", "hero", 3, 4))

		Expect(client.read(ctx).GetActionError().GetCode()).To(Equal(screensv1.ActionErrorCode_ACTION_ERROR_CODE_FAILED_PRECONDITION))
	})

	Describe("ScreenHandler.MoveEntity", func() {
		It("sends only the moved token to every listener, and keeps it in the state", func(ctx SpecContext) {
			handler := manager.GetOrCreate("tv")
			handler.SetState(ctx, battleMapState())
			client := connectScreen(ctx, manager, "tv", user)

			Expect(handler.MoveEntity(ctx, "hero", vectors.Vector2D{X: 5, Y: 6})).To(BeTrue())

			update := client.read(ctx).GetEntityUpdate()
			Expect(update.GetLayerId()).To(Equal("tokens"))
			Expect(update.GetEntity().GetId()).To(Equal("hero"))
			Expect(update.GetEntity().GetPosition().GetX()).To(Equal(5.0))
			Expect(update.GetEntity().GetPosition().GetY()).To(Equal(6.0))
			Expect(handler.State().GetLayer()[1].GetEntity()[0].GetPosition().GetX()).To(Equal(5.0), "a screen connecting later gets the new position")
		})

		It("sends nothing for entities the screen doesn't draw", func(ctx SpecContext) {
			handler := manager.GetOrCreate("tv")
			handler.SetState(ctx, battleMapState())

			Expect(handler.MoveEntity(ctx, "spy", vectors.Vector2D{X: 5, Y: 6})).To(BeFalse())
		})
	})

	Describe("ScreenHandler.SetCamera", func() {
		It("sends the camera to every listener, and keeps it in the state", func(ctx SpecContext) {
			handler := manager.GetOrCreate("tv")
			client := connectScreen(ctx, manager, "tv", user)

			handler.SetCamera(ctx, state.Camera{Center: vectors.Vector2D{X: 7, Y: 8}, Zoom: 3})

			camera := client.read(ctx).GetCameraUpdate()
			Expect(camera.GetCenter().GetX()).To(Equal(7.0))
			Expect(camera.GetZoom()).To(Equal(3.0))
			Expect(handler.State().GetCamera().GetZoom()).To(Equal(3.0))
		})
	})
})
