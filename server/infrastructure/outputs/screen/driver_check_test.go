package screen_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing/fstest"

	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/infrastructure/assets"
	"github.com/DaanV2/mechanus/server/infrastructure/outputs/screen"
	"github.com/coder/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// openScreen connects to the screen of screenID as the given connection and returns a function that closes it again.
func openScreen(ctx context.Context, manager *screens.ScreenManager, screenID string, info *screens.ConnectionInfo) func() {
	GinkgoHelper()
	connected := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer GinkgoRecover()

		conn, err := websocket.Accept(w, r, nil)
		Expect(err).ToNot(HaveOccurred())

		sc := screens.NewScreenConn("conn-"+info.ID, screenID, r.Context(), conn, info)
		manager.GetOrCreate(screenID).AddListener(r.Context(), sc) // forgets the listener once the connection ends
		close(connected)
		<-sc.Context().Done()
	}))
	DeferCleanup(server.Close)

	client, resp, err := websocket.Dial(ctx, server.URL, nil)
	Expect(err).ToNot(HaveOccurred())
	if resp.Body != nil {
		_ = resp.Body.Close()
	}
	DeferCleanup(func() { _ = client.CloseNow() })
	Eventually(connected).Should(BeClosed())

	return func() { _ = client.Close(websocket.StatusNormalClosure, "done") }
}

var _ = Describe("Driver.CheckDevice", func() {
	var (
		ctx     context.Context
		manager *screens.ScreenManager
		driver  *screen.Driver
		tv      = &devices.Device{ID: "table-tv", Kind: devices.KindScreen, Integration: devices.IntegrationScreen}
	)

	BeforeEach(func() {
		ctx = context.Background()
		manager = screens.NewScreenManager()
		driver = screen.NewDriver(manager, assets.NewStore(fstest.MapFS{}))
	})

	It("is offline while the device has no screen", func() {
		Expect(driver.CheckDevice(ctx, tv)).To(MatchError(screen.ErrNotConnected))
	})

	It("is offline when only a user looks at the device's screen", func() {
		manager.GetOrCreate(tv.ID)
		openScreen(ctx, manager, tv.ID, &screens.ConnectionInfo{ID: "user-1", Type: devices.DeviceTypeUser})

		Expect(driver.CheckDevice(ctx, tv)).To(MatchError(screen.ErrNotConnected))
	})

	It("is online while the device has its screen open, and offline once it closes", func() {
		closeScreen := openScreen(ctx, manager, tv.ID, &screens.ConnectionInfo{ID: tv.ID, Type: devices.DeviceTypeDevice, DeviceID: tv.ID})

		Expect(driver.CheckDevice(ctx, tv)).To(Succeed())

		closeScreen()
		Eventually(func() error { return driver.CheckDevice(ctx, tv) }).Should(MatchError(screen.ErrNotConnected))
	})
})
