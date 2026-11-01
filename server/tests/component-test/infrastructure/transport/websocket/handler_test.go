package websocket_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/websocket"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
)

var _ = Describe("WebsocketHandler authentication", func() {
	var (
		jwtService *authentication.JWTService
		router     http.Handler
	)

	BeforeEach(func(setupCtx SpecContext) {
		db := util_test.CreateDatabase(setupCtx)
		keyManager, err := authentication.NewKeyManager(storage.DBStorage[*authentication.KeyData](db))
		Expect(err).ShouldNot(HaveOccurred())
		jwtService = authentication.NewJWTService(authentication.NewJTIService(db), keyManager)

		handler := websocket.NewWebsocketHandler(screens.NewScreenManager(), jwtService, websocket.GetWebsocketConfig())
		router = websocket.NewWebsocketRouter(handler)
	})

	createToken := func(ctx context.Context, scope string) string {
		token, err := jwtService.Create(ctx, util_test.CreateUser(), authentication.GrantPassword, scope)
		Expect(err).ShouldNot(HaveOccurred())

		return token
	}

	// Sends a plain (non-upgrade) request: one that passes authentication gets to the upgrade and fails it with a 426, one that fails authentication gets a 401
	call := func(ctx context.Context, setHeaders func(h http.Header)) int {
		req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/screen/unknown/conn", nil)
		setHeaders(req.Header)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		return rec.Code
	}

	viaAuthorization := func(token string) func(h http.Header) {
		return func(h http.Header) { h.Set("Authorization", "Bearer "+token) }
	}
	viaSubprotocol := func(token string) func(h http.Header) {
		return func(h http.Header) {
			h.Set("Sec-WebSocket-Protocol", authentication.WebsocketSubprotocol+", "+authentication.WebsocketBearerSubprotocolPrefix+token)
		}
	}
	viaCookie := func(token string) func(h http.Header) {
		// Encoded the same way the browser client stores it
		return func(h http.Header) {
			h.Set("Cookie", authentication.AccessTokenCookie+"="+url.PathEscape("Bearer "+token))
		}
	}

	It("lets a connection without credentials through as a device", func(ctx SpecContext) {
		Expect(call(ctx, func(http.Header) {})).To(Equal(http.StatusUpgradeRequired))
	})

	DescribeTable("authenticates a valid user token",
		func(ctx SpecContext, via func(string) func(http.Header)) {
			Expect(call(ctx, via(createToken(ctx, authentication.ScopeAll)))).To(Equal(http.StatusUpgradeRequired))
		},
		Entry("from the Authorization header", viaAuthorization),
		Entry("from the subprotocol", viaSubprotocol),
		Entry("from the cookie", viaCookie),
	)

	DescribeTable("rejects an invalid user token",
		func(ctx SpecContext, via func(string) func(http.Header)) {
			Expect(call(ctx, via("not.a.jwt"))).To(Equal(http.StatusUnauthorized))
		},
		Entry("from the Authorization header", viaAuthorization),
		Entry("from the subprotocol", viaSubprotocol),
		Entry("from the cookie", viaCookie),
	)

	It("rejects a refresh-only token", func(ctx SpecContext) {
		Expect(call(ctx, viaCookie(createToken(ctx, authentication.ScopeRefresh)))).To(Equal(http.StatusUnauthorized))
	})
})
