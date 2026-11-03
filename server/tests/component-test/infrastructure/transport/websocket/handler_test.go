package websocket_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	cwebsocket "github.com/coder/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/websocket"
	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("WebsocketHandler", func() {
	var (
		jtiService    *authentication.JTIService
		jwtService    *authentication.JWTService
		screenManager *screens.ScreenManager
		config        websocket.WebsocketConfig
		router        http.Handler
	)

	setup := func() {
		handler := websocket.NewWebsocketHandler(screenManager, jwtService, config)
		router = websocket.NewWebsocketRouter(handler)
	}

	BeforeEach(func(setupCtx SpecContext) {
		db := util_test.CreateDatabase(setupCtx)
		keyManager, err := authentication.NewKeyManager(storage.DBStorage[*authentication.KeyData](db))
		Expect(err).ShouldNot(HaveOccurred())
		jtiService = authentication.NewJTIService(db)
		jwtService = authentication.NewJWTService(jtiService, keyManager)
		screenManager = screens.NewScreenManager()
		config = websocket.GetWebsocketConfig()
		setup()
	})

	createToken := func(ctx context.Context, scope string) string {
		token, err := jwtService.Create(ctx, util_test.CreateUser(), authentication.GrantPassword, scope)
		Expect(err).ShouldNot(HaveOccurred())

		return token
	}

	createDeviceTokens := func(ctx context.Context, deviceID string) *authentication.TokenPair {
		tokens, err := jwtService.CreateDeviceSession(ctx, deviceID, "Table TV")
		Expect(err).ShouldNot(HaveOccurred())

		return tokens
	}

	viaDeviceAuthorization := func(token string) func(h http.Header) {
		return func(h http.Header) { h.Set("Authorization", authentication.DeviceTokenType+" "+token) }
	}
	viaDeviceSubprotocol := func(token string) func(h http.Header) {
		return func(h http.Header) {
			h.Set("Sec-WebSocket-Protocol", authentication.WebsocketSubprotocol+", "+authentication.WebsocketDeviceSubprotocolPrefix+token)
		}
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

	Describe("authentication", func() {
		// Sends a plain (non-upgrade) request: one that passes authentication and the screen access check gets
		// to the upgrade and fails it with a 426, others get a 401 or 403
		call := func(ctx context.Context, path string, setHeaders func(h http.Header)) int {
			req := httptest.NewRequestWithContext(ctx, http.MethodGet, path, nil)
			setHeaders(req.Header)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			return rec.Code
		}

		It("answers a 401 to a connection without credentials, as the route needs an identity", func(ctx SpecContext) {
			Expect(call(ctx, "/api/v1/screen/tv-1/conn?deviceId=tv-1", func(http.Header) {})).To(Equal(http.StatusUnauthorized))
		})

		DescribeTable("authenticates a paired device, to its own screen",
			func(ctx SpecContext, via func(string) func(http.Header)) {
				Expect(call(ctx, "/api/v1/screen/tv-1/conn?deviceId=tv-1", via(createDeviceTokens(ctx, "tv-1").Access))).To(Equal(http.StatusUpgradeRequired))
			},
			Entry("from the Authorization header", viaDeviceAuthorization),
			Entry("from the subprotocol", viaDeviceSubprotocol),
		)

		It("takes the device ID from the device token when the device reports none", func(ctx SpecContext) {
			Expect(call(ctx, "/api/v1/screen/tv-1/conn", viaDeviceAuthorization(createDeviceTokens(ctx, "tv-1").Access))).To(Equal(http.StatusUpgradeRequired))
		})

		It("rejects a device reporting another ID than its token", func(ctx SpecContext) {
			Expect(call(ctx, "/api/v1/screen/tv-1/conn?deviceId=tv-2", viaDeviceAuthorization(createDeviceTokens(ctx, "tv-1").Access))).To(Equal(http.StatusUnauthorized))
		})

		DescribeTable("rejects an invalid device token",
			func(ctx SpecContext, via func(string) func(http.Header)) {
				Expect(call(ctx, "/api/v1/screen/tv-1/conn?deviceId=tv-1", via("not.a.jwt"))).To(Equal(http.StatusUnauthorized))
			},
			Entry("from the Authorization header", viaDeviceAuthorization),
			Entry("from the subprotocol", viaDeviceSubprotocol),
		)

		It("rejects a revoked device token instead of treating it as anonymous", func(ctx SpecContext) {
			tokens := createDeviceTokens(ctx, "tv-1")
			revoked, err := jtiService.Revoke(ctx, tokens.JTI)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(revoked).To(BeTrue())

			Expect(call(ctx, "/api/v1/screen/tv-1/conn", viaDeviceAuthorization(tokens.Access))).To(Equal(http.StatusUnauthorized))
		})

		It("rejects a device refresh token", func(ctx SpecContext) {
			Expect(call(ctx, "/api/v1/screen/tv-1/conn", viaDeviceAuthorization(createDeviceTokens(ctx, "tv-1").Refresh))).To(Equal(http.StatusUnauthorized))
		})

		It("rejects a device token passed as a user token", func(ctx SpecContext) {
			Expect(call(ctx, "/api/v1/screen/tv-1/conn", viaAuthorization(createDeviceTokens(ctx, "tv-1").Access))).To(Equal(http.StatusUnauthorized))
		})

		It("rejects a user token passed as a device token", func(ctx SpecContext) {
			Expect(call(ctx, "/api/v1/screen/tv-1/conn", viaDeviceAuthorization(createToken(ctx, authentication.ScopeAll)))).To(Equal(http.StatusUnauthorized))
		})

		It("rejects an unknown authorization scheme", func(ctx SpecContext) {
			Expect(call(ctx, "/api/v1/screen/tv-1/conn", func(h http.Header) { h.Set("Authorization", "Basic abc") })).To(Equal(http.StatusUnauthorized))
		})

		DescribeTable("authenticates a valid user token",
			func(ctx SpecContext, via func(string) func(http.Header)) {
				Expect(call(ctx, "/api/v1/screen/unknown/conn", via(createToken(ctx, authentication.ScopeAll)))).To(Equal(http.StatusUpgradeRequired))
			},
			Entry("from the Authorization header", viaAuthorization),
			Entry("from the subprotocol", viaSubprotocol),
			Entry("from the cookie", viaCookie),
		)

		DescribeTable("rejects an invalid user token",
			func(ctx SpecContext, via func(string) func(http.Header)) {
				Expect(call(ctx, "/api/v1/screen/unknown/conn", via("not.a.jwt"))).To(Equal(http.StatusUnauthorized))
			},
			Entry("from the Authorization header", viaAuthorization),
			Entry("from the subprotocol", viaSubprotocol),
			Entry("from the cookie", viaCookie),
		)

		It("rejects a refresh-only token", func(ctx SpecContext) {
			Expect(call(ctx, "/api/v1/screen/unknown/conn", viaCookie(createToken(ctx, authentication.ScopeRefresh)))).To(Equal(http.StatusUnauthorized))
		})

		It("rejects a revoked session instead of treating it as a device", func(ctx SpecContext) {
			tokens, err := jwtService.CreateSession(ctx, util_test.CreateUser(), authentication.GrantPassword)
			Expect(err).ShouldNot(HaveOccurred())
			revokeSession(ctx, jwtService, jtiService, tokens.Access)

			Expect(call(ctx, "/api/v1/screen/tv-1/conn?deviceId=tv-1", viaAuthorization(tokens.Access))).To(Equal(http.StatusUnauthorized))
		})

		It("forbids a device to open another screen", func(ctx SpecContext) {
			Expect(call(ctx, "/api/v1/screen/tv-1/conn", viaDeviceAuthorization(createDeviceTokens(ctx, "tv-2").Access))).To(Equal(http.StatusForbidden))
		})

		It("lets a user open any screen", func(ctx SpecContext) {
			Expect(call(ctx, "/api/v1/screen/tv-1/conn?deviceId=tv-2", viaAuthorization(createToken(ctx, authentication.ScopeAll)))).To(Equal(http.StatusUpgradeRequired))
		})

		It("requires a screen and connection ID", func(ctx SpecContext) {
			Expect(call(ctx, "/api/v1/screen/tv-1/", viaAuthorization(createToken(ctx, authentication.ScopeAll)))).To(Equal(http.StatusNotFound))
		})
	})

	Describe("connections", func() {
		var server *httptest.Server

		JustBeforeEach(func() {
			server = httptest.NewServer(router)
			DeferCleanup(server.Close)
		})

		dial := func(ctx context.Context, path string, setHeaders func(h http.Header)) (*cwebsocket.Conn, error) {
			header := http.Header{}
			setHeaders(header)

			conn, _, err := cwebsocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+path, &cwebsocket.DialOptions{HTTPHeader: header})
			if conn != nil {
				DeferCleanup(func() { _ = conn.CloseNow() })
			}

			return conn, err
		}

		// connectionsOf returns the connections the server has registered on the screen
		connectionsOf := func(screenID string) func() []*screens.ConnectionInfo {
			return func() []*screens.ConnectionInfo {
				screen, ok := screenManager.Get(screenID)
				if !ok {
					return nil
				}

				return screen.Connections()
			}
		}

		// closeStatus reads from the connection until the server closes it, and returns the close status
		closeStatus := func(ctx context.Context, conn *cwebsocket.Conn) cwebsocket.StatusCode {
			for {
				if _, _, err := conn.Read(ctx); err != nil {
					return cwebsocket.CloseStatus(err)
				}
			}
		}

		DescribeTable("stores the device identity the client reports",
			func(ctx SpecContext, path string, headers map[string]string) {
				_, err := dial(ctx, path, func(h http.Header) {
					viaDeviceAuthorization(createDeviceTokens(ctx, "tv-1").Access)(h)
					for k, v := range headers {
						h.Set(k, v)
					}
				})
				Expect(err).ShouldNot(HaveOccurred())

				Eventually(connectionsOf("tv-1")).Should(ConsistOf(And(
					HaveField("DeviceID", "tv-1"),
					HaveField("Hardware", devices.HardwareTable),
					HaveField("Type", devices.DeviceTypeDevice),
				)))
			},
			Entry("from the headers", "/api/v1/screen/tv-1/conn", map[string]string{
				websocket.DeviceIDHeader: "tv-1", websocket.DeviceTypeHeader: "table",
			}),
			Entry("from the query", "/api/v1/screen/tv-1/conn?deviceId=tv-1&deviceType=table", nil),
			Entry("preferring the headers over the query", "/api/v1/screen/tv-1/conn?deviceId=tv-2&deviceType=phone", map[string]string{
				websocket.DeviceIDHeader: "tv-1", websocket.DeviceTypeHeader: "table",
			}),
		)

		It("stays open for a device", func(ctx SpecContext) {
			conn, err := dial(ctx, "/api/v1/screen/tv-1/conn?deviceId=tv-1", viaDeviceAuthorization(createDeviceTokens(ctx, "tv-1").Access))
			Expect(err).ShouldNot(HaveOccurred())

			readCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			defer cancel()
			_, _, err = conn.Read(readCtx)
			Expect(err).To(MatchError(context.DeadlineExceeded))
			Expect(connectionsOf("tv-1")()).To(HaveLen(1))
		})

		It("answers every message on the connection", func(ctx SpecContext) {
			conn, err := dial(ctx, "/api/v1/screen/tv-1/conn?deviceId=tv-1", viaDeviceAuthorization(createDeviceTokens(ctx, "tv-1").Access))
			Expect(err).ShouldNot(HaveOccurred())

			request, err := proto.Marshal(&screensv1.ClientMessages{Action: []*screensv1.ClientMessage{{
				Action: &screensv1.ClientMessage_InitialSetupRequest{InitialSetupRequest: &screensv1.InitialSetupRequest{}},
			}}})
			Expect(err).ShouldNot(HaveOccurred())

			for range 2 {
				Expect(conn.Write(ctx, cwebsocket.MessageBinary, request)).To(Succeed())

				_, data, err := conn.Read(ctx)
				Expect(err).ShouldNot(HaveOccurred())
				var msg screensv1.ServerMessages
				Expect(proto.Unmarshal(data, &msg)).To(Succeed())
				Expect(msg.GetAction()).To(ContainElement(WithTransform((*screensv1.ServerMessage).GetInitialSetup, Not(BeNil()))))
			}
		}, SpecTimeout(10*time.Second))

		It("answers a 403 to a device opening another screen", func(ctx SpecContext) {
			_, err := dial(ctx, "/api/v1/screen/tv-1/conn?deviceId=tv-2", viaDeviceAuthorization(createDeviceTokens(ctx, "tv-2").Access))
			Expect(err).To(MatchError(ContainSubstring("403")))
			Expect(connectionsOf("tv-1")()).To(BeEmpty())
		})

		It("answers a 401 to a connection without credentials, leaving no screen behind", func(ctx SpecContext) {
			_, err := dial(ctx, "/api/v1/screen/tv-1/conn?deviceId=tv-1", func(http.Header) {})
			Expect(err).To(MatchError(ContainSubstring("401")))

			_, ok := screenManager.Get("tv-1")
			Expect(ok).To(BeFalse())
		})

		It("closes a device connection once it gets revoked", func(ctx SpecContext) {
			config.SessionCheckInterval = 50 * time.Millisecond
			setup()
			server.Config.Handler = router

			tokens := createDeviceTokens(ctx, "tv-1")
			conn, err := dial(ctx, "/api/v1/screen/tv-1/conn", viaDeviceAuthorization(tokens.Access))
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(connectionsOf("tv-1")).Should(HaveLen(1))

			_, err = jtiService.Revoke(ctx, tokens.JTI)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(closeStatus(ctx, conn)).To(Equal(cwebsocket.StatusPolicyViolation))
		}, SpecTimeout(10*time.Second))

		It("closes the connections of a device on request, leaving users on its screen connected", func(ctx SpecContext) {
			deviceConn, err := dial(ctx, "/api/v1/screen/tv-1/device", viaDeviceAuthorization(createDeviceTokens(ctx, "tv-1").Access))
			Expect(err).ShouldNot(HaveOccurred())
			_, err = dial(ctx, "/api/v1/screen/tv-1/user", viaAuthorization(createToken(ctx, authentication.ScopeAll)))
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(connectionsOf("tv-1")).Should(HaveLen(2))

			screenManager.CloseDevice("tv-1", "device revoked")

			Expect(closeStatus(ctx, deviceConn)).To(Equal(cwebsocket.StatusPolicyViolation))
			Eventually(connectionsOf("tv-1")).Should(ConsistOf(HaveField("Type", devices.DeviceTypeUser)))
		}, SpecTimeout(10*time.Second))

		It("closes a device connection once its token expires", func(ctx SpecContext) {
			jwtService = jwtService.WithOptions(authentication.JWTOptions{DeviceTokenDuration: time.Second, DeviceRefreshTokenDuration: time.Hour})
			setup()
			server.Config.Handler = router

			conn, err := dial(ctx, "/api/v1/screen/tv-1/conn", viaDeviceAuthorization(createDeviceTokens(ctx, "tv-1").Access))
			Expect(err).ShouldNot(HaveOccurred())

			Expect(closeStatus(ctx, conn)).To(Equal(cwebsocket.StatusPolicyViolation))
		}, SpecTimeout(10*time.Second))

		It("closes a user connection once its token expires", func(ctx SpecContext) {
			jwtService = jwtService.WithOptions(authentication.JWTOptions{TokenDuration: time.Second, RefreshTokenDuration: time.Hour})
			setup()
			server.Config.Handler = router

			conn, err := dial(ctx, "/api/v1/screen/tv-1/conn", viaAuthorization(createToken(ctx, authentication.ScopeAll)))
			Expect(err).ShouldNot(HaveOccurred())

			Expect(closeStatus(ctx, conn)).To(Equal(cwebsocket.StatusPolicyViolation))
		}, SpecTimeout(10*time.Second))

		It("closes a user connection once its session is revoked", func(ctx SpecContext) {
			config.SessionCheckInterval = 50 * time.Millisecond
			setup()
			server.Config.Handler = router

			tokens, err := jwtService.CreateSession(ctx, util_test.CreateUser(), authentication.GrantPassword)
			Expect(err).ShouldNot(HaveOccurred())
			conn, err := dial(ctx, "/api/v1/screen/tv-1/conn", viaCookie(tokens.Access))
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(connectionsOf("tv-1")).Should(HaveLen(1))

			revokeSession(ctx, jwtService, jtiService, tokens.Access)

			Expect(closeStatus(ctx, conn)).To(Equal(cwebsocket.StatusPolicyViolation))
		}, SpecTimeout(10*time.Second))
	})
})

// revokeSession revokes the session (JTI) the token belongs to.
func revokeSession(ctx context.Context, jwts *authentication.JWTService, jtis *authentication.JTIService, token string) {
	GinkgoHelper()

	t, err := jwts.Validate(ctx, token)
	Expect(err).ShouldNot(HaveOccurred())
	claims, ok := authentication.GetClaims(t.Claims)
	Expect(ok).To(BeTrue())

	revoked, err := jtis.Revoke(ctx, claims.ID)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(revoked).To(BeTrue())
}
