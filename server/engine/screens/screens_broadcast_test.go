package screens_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"

	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/engine/screens/state"
	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
	"github.com/coder/websocket"
)

var _ = Describe("ScreenHandler.SetState", func() {
	It("reaches listeners when the caller's context has already ended", func(ctx SpecContext) {
		handler := screens.NewScreenHandler("tv")
		connected := make(chan struct{})

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()

			conn, err := websocket.Accept(w, r, nil)
			Expect(err).ToNot(HaveOccurred())

			info := &screens.ConnectionInfo{}
			handler.AddListener(context.Background(), screens.NewScreenConn("conn", "tv", context.Background(), conn, info))
			close(connected)
		}))
		DeferCleanup(server.Close)

		client, resp, err := websocket.Dial(ctx, server.URL, nil)
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(func() {
			if resp.Body != nil {
				_ = resp.Body.Close()
			}
		})
		DeferCleanup(func() { _ = client.CloseNow() })
		Eventually(connected).Should(BeClosed())

		// A scene change comes from a request, which is over before the write to the screen happens
		ended, cancel := context.WithCancel(ctx)
		cancel()

		st := state.NewScreenState()
		st.View.Kind = scenes.ViewBattleMap
		handler.SetState(ended, st)

		readCtx, stop := context.WithTimeout(ctx, 5*time.Second)
		defer stop()
		_, data, err := client.Read(readCtx)
		Expect(err).ToNot(HaveOccurred())

		var msgs screensv1.ServerMessages
		Expect(proto.Unmarshal(data, &msgs)).To(Succeed())
		Expect(msgs.GetAction()).To(HaveLen(1))
		Expect(msgs.GetAction()[0].GetScreenState().GetView().GetKind()).To(Equal(screensv1.ViewKind_VIEW_KIND_BATTLEMAP))
	})
})
