package grpc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"
	devicesv1 "github.com/DaanV2/mechanus/server/proto/devices/v1"
	"github.com/DaanV2/mechanus/server/proto/devices/v1/devicesv1connect"
	usersv1 "github.com/DaanV2/mechanus/server/proto/users/v1"
)

// pairing is the device's side of an open DeviceService.Pair stream.
type pairing struct {
	cancel context.CancelFunc
	events chan *devicesv1.PairResponse
	result chan error
}

// nextCode returns the next code the device is told to show.
func (p *pairing) nextCode() string {
	GinkgoHelper()

	var event *devicesv1.PairResponse
	Eventually(p.events, 5*time.Second).Should(Receive(&event))
	Expect(event.GetCode()).ToNot(BeNil())
	Expect(event.GetCode().GetCode()).To(MatchRegexp(`^\d{4}$`))

	return event.GetCode().GetCode()
}

var _ = Describe("DeviceService", func() {
	var (
		env     *authEnv
		service *application.DeviceService
		manager *screens.ScreenManager
		client  devicesv1connect.DeviceServiceClient
		admin   string
	)

	start := func(options application.PairingOptions) {
		service = application.NewDeviceService(repositories.NewDeviceRepository(env.db), env.jwts, env.jtis, manager).WithPairingOptions(options)
		_, handler := devicesv1connect.NewDeviceServiceHandler(
			grpc.NewDeviceServiceHandler(service),
			connect.WithInterceptors(grpc.NewPolicyInterceptor(authentication.NewIdentityResolver(env.jwts), grpc.DeviceServicePolicies())),
		)
		server := httptest.NewServer(handler)
		DeferCleanup(server.Close)

		client = devicesv1connect.NewDeviceServiceClient(server.Client(), server.URL, authorizationFromContext())
	}

	asUser := func(ctx context.Context, role roles.Role) context.Context {
		GinkgoHelper()

		tokens, err := env.jwts.CreateSession(ctx, env.createUser(ctx, "password", role.String()), authentication.GrantPassword)
		Expect(err).ShouldNot(HaveOccurred())

		return withAuthorization(ctx, "Bearer "+tokens.Access)
	}

	// pair opens a pairing stream as the device, it stays open until the device is paired or the spec ends
	pair := func(ctx context.Context, deviceID, hardware string) *pairing {
		GinkgoHelper()

		streamCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
		DeferCleanup(cancel)

		req := connect.NewRequest(&devicesv1.PairRequest{})
		req.Header().Set(authentication.DeviceIDHeader, deviceID)
		req.Header().Set(authentication.DeviceTypeHeader, hardware)
		stream, err := client.Pair(streamCtx, req)
		Expect(err).ShouldNot(HaveOccurred())

		p := &pairing{cancel: cancel, events: make(chan *devicesv1.PairResponse, 8), result: make(chan error, 1)}
		go func() {
			for stream.Receive() {
				p.events <- stream.Msg()
			}
			p.result <- stream.Err()
		}()

		return p
	}

	confirm := func(ctx context.Context, code, name string) (*connect.Response[devicesv1.ConfirmPairingResponse], error) {
		return client.Confirm(withAuthorization(ctx, admin), connect.NewRequest(&devicesv1.ConfirmPairingRequest{Code: code, Name: name}))
	}

	// pairDevice runs the whole flow and returns the tokens the device ends up with
	pairDevice := func(ctx context.Context, deviceID string) *devicesv1.DeviceTokens {
		GinkgoHelper()

		p := pair(ctx, deviceID, "table")
		_, err := confirm(ctx, p.nextCode(), "Table TV")
		Expect(err).ShouldNot(HaveOccurred())

		var event *devicesv1.PairResponse
		Eventually(p.events, 5*time.Second).Should(Receive(&event))
		Expect(event.GetTokens()).ToNot(BeNil())
		Eventually(p.result, 5*time.Second).Should(Receive(BeNil()))

		return event.GetTokens()
	}

	list := func(ctx context.Context) []*devicesv1.Device {
		GinkgoHelper()

		resp, err := client.List(withAuthorization(ctx, admin), connect.NewRequest(&devicesv1.ListDevicesRequest{}))
		Expect(err).ShouldNot(HaveOccurred())

		return resp.Msg.GetDevices()
	}

	BeforeEach(func(setupCtx SpecContext) {
		env = newAuthEnv(setupCtx)
		manager = screens.NewScreenManager()
		start(application.DefaultPairingOptions())

		tokens, err := env.jwts.CreateSession(setupCtx, env.createUser(setupCtx, "password", roles.Admin.String()), authentication.GrantPassword)
		Expect(err).ShouldNot(HaveOccurred())
		admin = "Bearer " + tokens.Access
	})

	Describe("the pairing flow", func() {
		It("ends with device tokens for the device that showed the code", func(ctx SpecContext) {
			tokens := pairDevice(ctx, "tv-1")
			Expect(tokens.GetType()).To(Equal("Device"))

			access := env.claims(ctx, tokens.GetToken())
			Expect(access.Subject).To(Equal("tv-1"))
			Expect(access.Scope).To(Equal(authentication.ScopeDevice))
			Expect(access.Grant).To(Equal(authentication.GrantPairing))
			Expect(access.User.Roles).To(ConsistOf(roles.Device.String()))
			Expect(time.Until(access.ExpiresAt.Time)).To(BeNumerically("~", 8*time.Hour, time.Minute))

			refresh := env.claims(ctx, tokens.GetRefreshToken())
			Expect(refresh.Subject).To(Equal("tv-1"))
			Expect(refresh.Scope).To(Equal(authentication.ScopeDeviceRefresh))
			Expect(refresh.ID).To(Equal(access.ID))
			Expect(time.Until(refresh.ExpiresAt.Time)).To(BeNumerically("~", 30*24*time.Hour, time.Minute))
		}, SpecTimeout(15*time.Second))

		It("registers the device under its stable ID", func(ctx SpecContext) {
			pairDevice(ctx, "tv-1")

			Expect(list(ctx)).To(ConsistOf(And(
				WithTransform((*devicesv1.Device).GetId, Equal("tv-1")),
				WithTransform((*devicesv1.Device).GetName, Equal("Table TV")),
				WithTransform((*devicesv1.Device).GetHardware, Equal("table")),
				WithTransform((*devicesv1.Device).GetPaired, BeTrue()),
				WithTransform((*devicesv1.Device).GetConnected, BeFalse()),
			)))
		}, SpecTimeout(15*time.Second))

		It("needs a device ID", func(ctx SpecContext) {
			p := pair(ctx, "", "screen")
			var err error
			Eventually(p.result, 5*time.Second).Should(Receive(&err))
			Expect(err).To(HaveConnectCode(connect.CodeInvalidArgument))
		}, SpecTimeout(15*time.Second))

		DescribeTable("refuses an unusable device ID",
			func(ctx SpecContext, id string) {
				p := pair(ctx, id, "screen")
				var err error
				Eventually(p.result, 5*time.Second).Should(Receive(&err))
				Expect(err).To(HaveConnectCode(connect.CodeInvalidArgument))
			},
			Entry("path separators", "../tv"),
			Entry("spaces", "tv 1"),
		)

		It("rejects an invalid credential, never treating it as anonymous", func(ctx SpecContext) {
			req := connect.NewRequest(&devicesv1.PairRequest{})
			req.Header().Set(authentication.DeviceIDHeader, "tv-1")
			req.Header().Set("Authorization", "Device not.a.jwt")
			stream, err := client.Pair(ctx, req)
			if err == nil {
				Expect(stream.Receive()).To(BeFalse())
				err = stream.Err()
			}
			Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))
		}, SpecTimeout(15*time.Second))

		It("needs a name", func(ctx SpecContext) {
			p := pair(ctx, "tv-1", "screen")
			_, err := confirm(ctx, p.nextCode(), "   ")
			Expect(err).To(HaveConnectCode(connect.CodeInvalidArgument))
		}, SpecTimeout(15*time.Second))
	})

	Describe("the code", func() {
		It("is single use", func(ctx SpecContext) {
			p := pair(ctx, "tv-1", "screen")
			code := p.nextCode()
			_, err := confirm(ctx, code, "Table TV")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = confirm(ctx, code, "Table TV")
			Expect(err).To(HaveConnectCode(connect.CodeFailedPrecondition))
		}, SpecTimeout(15*time.Second))

		It("stops being valid once the stream closes", func(ctx SpecContext) {
			streamCtx, cancel := context.WithCancel(ctx)
			defer cancel()

			codes := make(chan string, 1)
			ended := make(chan error, 1)
			go func() {
				ended <- service.Pair(streamCtx, "tv-1", devices.HardwareScreen, func(event application.PairingEvent) error {
					codes <- event.Code

					return nil
				})
			}()

			var code string
			Eventually(codes, 5*time.Second).Should(Receive(&code))
			cancel()
			Eventually(ended, 5*time.Second).Should(Receive(MatchError(context.Canceled)))

			_, err := confirm(ctx, code, "Table TV")
			Expect(err).To(HaveConnectCode(connect.CodeFailedPrecondition))
			Expect(list(ctx)).To(BeEmpty())
		}, SpecTimeout(15*time.Second))

		It("gets renewed when it expires, the old one stops working", func(ctx SpecContext) {
			options := application.DefaultPairingOptions()
			options.CodeLifetime = 100 * time.Millisecond
			start(options)

			p := pair(ctx, "tv-1", "screen")
			first := p.nextCode()
			latest := p.nextCode()
			for latest == first { // Codes are random, the same one can come up twice
				latest = p.nextCode()
			}

			_, err := confirm(ctx, first, "Table TV")
			Expect(err).To(HaveConnectCode(connect.CodeFailedPrecondition))
		}, SpecTimeout(15*time.Second))

		It("of a device starting over replaces the old stream", func(ctx SpecContext) {
			first := pair(ctx, "tv-1", "screen")
			first.nextCode()
			second := pair(ctx, "tv-1", "screen")
			second.nextCode()

			var err error
			Eventually(first.result, 5*time.Second).Should(Receive(&err))
			Expect(err).To(HaveConnectCode(connect.CodeAborted))
		}, SpecTimeout(15*time.Second))
	})

	Describe("rate limiting", func() {
		wrongCode := func(code string) string {
			if code == "0000" {
				return "0001"
			}

			return "0000"
		}

		It("replaces every open code after a few wrong guesses, and refuses more attempts", func(ctx SpecContext) {
			options := application.DefaultPairingOptions()
			options.MaxWrongCodes = 3
			start(options)

			p := pair(ctx, "tv-1", "screen")
			code := p.nextCode()

			for range 3 {
				_, err := confirm(ctx, wrongCode(code), "Table TV")
				Expect(err).To(HaveConnectCode(connect.CodeFailedPrecondition))
			}

			// The device got a new code, as a guess can't be tied to one code
			renewed := p.nextCode()

			_, err := confirm(ctx, renewed, "Table TV")
			Expect(err).To(HaveConnectCode(connect.CodeResourceExhausted))
			Expect(list(ctx)).To(BeEmpty())
		}, SpecTimeout(15*time.Second))

		It("lets attempts through again after the window", func(ctx SpecContext) {
			options := application.DefaultPairingOptions()
			options.MaxWrongCodes = 1
			options.GuessWindow = 100 * time.Millisecond
			start(options)

			p := pair(ctx, "tv-1", "screen")
			code := p.nextCode()
			_, err := confirm(ctx, wrongCode(code), "Table TV")
			Expect(err).To(HaveConnectCode(connect.CodeFailedPrecondition))
			renewed := p.nextCode()

			Eventually(func() error {
				_, err := confirm(ctx, renewed, "Table TV")

				return err
			}, 5*time.Second, 50*time.Millisecond).Should(Succeed())
		}, SpecTimeout(15*time.Second))

		It("limits how many devices pair at once", func(ctx SpecContext) {
			options := application.DefaultPairingOptions()
			options.MaxSessions = 1
			start(options)

			pair(ctx, "tv-1", "screen").nextCode()

			p := pair(ctx, "tv-2", "screen")
			var err error
			Eventually(p.result, 5*time.Second).Should(Receive(&err))
			Expect(err).To(HaveConnectCode(connect.CodeResourceExhausted))
		}, SpecTimeout(15*time.Second))
	})

	Describe("policies", func() {
		adminOnly := map[string]func(ctx context.Context) error{
			"Confirm": func(ctx context.Context) error {
				_, err := client.Confirm(ctx, connect.NewRequest(&devicesv1.ConfirmPairingRequest{Code: "0000", Name: "TV"}))

				return err
			},
			"List": func(ctx context.Context) error {
				_, err := client.List(ctx, connect.NewRequest(&devicesv1.ListDevicesRequest{}))

				return err
			},
			"Revoke": func(ctx context.Context) error {
				_, err := client.Revoke(ctx, connect.NewRequest(&devicesv1.RevokeDeviceRequest{Id: "tv-1"}))

				return err
			},
			"Bind": func(ctx context.Context) error {
				_, err := client.Bind(ctx, connect.NewRequest(&devicesv1.BindDeviceRequest{Id: "tv-1", Target: "table"}))

				return err
			},
		}

		for name, call := range adminOnly {
			Describe(name, func() {
				It("answers Unauthenticated without a credential", func(ctx SpecContext) {
					Expect(call(ctx)).To(HaveConnectCode(connect.CodeUnauthenticated))
				})

				It("answers Unauthenticated to an invalid credential", func(ctx SpecContext) {
					Expect(call(withAuthorization(ctx, "Bearer not.a.jwt"))).To(HaveConnectCode(connect.CodeUnauthenticated))
				})

				It("answers PermissionDenied to a user without the role", func(ctx SpecContext) {
					Expect(call(asUser(ctx, roles.User))).To(HaveConnectCode(connect.CodePermissionDenied))
				})

				It("answers PermissionDenied to a paired device", func(ctx SpecContext) {
					tokens := pairDevice(ctx, "tv-2")
					Expect(call(withAuthorization(ctx, "Device "+tokens.GetToken()))).To(HaveConnectCode(connect.CodePermissionDenied))
				}, SpecTimeout(15*time.Second))
			})
		}

		It("denies a procedure without a policy", func(ctx SpecContext) {
			interceptor := grpc.NewPolicyInterceptor(authentication.NewIdentityResolver(env.jwts), map[string]authentication.Policy{})
			_, handler := devicesv1connect.NewDeviceServiceHandler(grpc.NewDeviceServiceHandler(service), connect.WithInterceptors(interceptor))
			server := httptest.NewServer(handler)
			DeferCleanup(server.Close)

			uncovered := devicesv1connect.NewDeviceServiceClient(http.DefaultClient, server.URL, authorizationFromContext())
			_, err := uncovered.List(withAuthorization(ctx, admin), connect.NewRequest(&devicesv1.ListDevicesRequest{}))
			Expect(err).To(HaveConnectCode(connect.CodePermissionDenied))
		})

		It("covers every procedure of the service", func() {
			Expect(grpc.DeviceServicePolicies()).To(HaveLen(6))
		})
	})

	Describe("Refresh", func() {
		refresh := func(ctx context.Context, token string) (*devicesv1.DeviceTokens, error) {
			resp, err := client.Refresh(ctx, connect.NewRequest(&devicesv1.RefreshDeviceRequest{Token: token}))
			if err != nil {
				return nil, err
			}

			return resp.Msg.GetTokens(), nil
		}

		It("gives new tokens of the same session", func(ctx SpecContext) {
			tokens := pairDevice(ctx, "tv-1")

			renewed, err := refresh(ctx, tokens.GetRefreshToken())
			Expect(err).ShouldNot(HaveOccurred())
			claims := env.claims(ctx, renewed.GetToken())
			Expect(claims.ID).To(Equal(env.claims(ctx, tokens.GetToken()).ID))
			Expect(claims.Scope).To(Equal(authentication.ScopeDevice))
			Expect(claims.Grant).To(Equal(authentication.GrantRefresh))
			Expect(claims.User.Name).To(Equal("Table TV"))
		}, SpecTimeout(15*time.Second))

		It("refuses an access token", func(ctx SpecContext) {
			_, err := refresh(ctx, pairDevice(ctx, "tv-1").GetToken())
			Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))
		}, SpecTimeout(15*time.Second))

		It("refuses the refresh token of a user", func(ctx SpecContext) {
			tokens, err := env.jwts.CreateSession(ctx, env.createUser(ctx, "password", roles.User.String()), authentication.GrantPassword)
			Expect(err).ShouldNot(HaveOccurred())

			_, err = refresh(ctx, tokens.Refresh)
			Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))
		})

		It("refuses a device refresh token at the login service", func(ctx SpecContext) {
			tokens := pairDevice(ctx, "tv-1")

			login := grpc.NewLoginServiceHandler(env.users, env.jwts, env.jtis)
			_, err := login.Refresh(ctx, connect.NewRequest(&usersv1.RefreshTokenRequest{Token: tokens.GetRefreshToken()}))
			Expect(err).To(HaveConnectCode(connect.CodePermissionDenied))
		}, SpecTimeout(15*time.Second))
	})

	Describe("Revoke", func() {
		It("un-pairs the device: its tokens stop working right away", func(ctx SpecContext) {
			tokens := pairDevice(ctx, "tv-1")

			_, err := client.Revoke(withAuthorization(ctx, admin), connect.NewRequest(&devicesv1.RevokeDeviceRequest{Id: "tv-1"}))
			Expect(err).ShouldNot(HaveOccurred())

			Expect(list(ctx)).To(BeEmpty())
			_, err = authentication.NewIdentityResolver(env.jwts).Resolve(ctx, http.Header{"Authorization": {"Device " + tokens.GetToken()}})
			Expect(err).To(MatchError(authentication.ErrInvalidCredential))

			_, err = client.Refresh(ctx, connect.NewRequest(&devicesv1.RefreshDeviceRequest{Token: tokens.GetRefreshToken()}))
			Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))
		}, SpecTimeout(15*time.Second))

		It("lets the device pair again", func(ctx SpecContext) {
			pairDevice(ctx, "tv-1")
			_, err := client.Revoke(withAuthorization(ctx, admin), connect.NewRequest(&devicesv1.RevokeDeviceRequest{Id: "tv-1"}))
			Expect(err).ShouldNot(HaveOccurred())

			pairDevice(ctx, "tv-1")
			Expect(list(ctx)).To(HaveLen(1))
		}, SpecTimeout(15*time.Second))

		It("answers NotFound for an unknown device", func(ctx SpecContext) {
			_, err := client.Revoke(withAuthorization(ctx, admin), connect.NewRequest(&devicesv1.RevokeDeviceRequest{Id: "tv-1"}))
			Expect(err).To(HaveConnectCode(connect.CodeNotFound))
		})

		It("only touches the session of the device, not a user sharing its ID", func(ctx SpecContext) {
			user := env.createUser(ctx, "password", roles.Admin.String())
			session, err := env.jwts.CreateSession(ctx, user, authentication.GrantPassword)
			Expect(err).ShouldNot(HaveOccurred())

			pairDevice(ctx, user.ID)
			_, err = client.Revoke(withAuthorization(ctx, admin), connect.NewRequest(&devicesv1.RevokeDeviceRequest{Id: user.ID}))
			Expect(err).ShouldNot(HaveOccurred())

			Expect(env.isRevoked(ctx, session.JTI)).To(BeFalse())
		}, SpecTimeout(15*time.Second))
	})

	Describe("pairing again", func() {
		It("revokes the old tokens of the device, and keeps its target", func(ctx SpecContext) {
			oldSession := env.claims(ctx, pairDevice(ctx, "tv-1").GetToken()).ID
			_, err := client.Bind(withAuthorization(ctx, admin), connect.NewRequest(&devicesv1.BindDeviceRequest{Id: "tv-1", Target: "table"}))
			Expect(err).ShouldNot(HaveOccurred())

			pairDevice(ctx, "tv-1")

			Expect(env.isRevoked(ctx, oldSession)).To(BeTrue())
			Expect(list(ctx)).To(ConsistOf(WithTransform((*devicesv1.Device).GetTarget, Equal("table"))))
		}, SpecTimeout(15*time.Second))
	})

	Describe("Bind", func() {
		bind := func(ctx context.Context, id, target string) error {
			_, err := client.Bind(withAuthorization(ctx, admin), connect.NewRequest(&devicesv1.BindDeviceRequest{Id: id, Target: target}))

			return err
		}

		It("binds a paired device to a target of the room", func(ctx SpecContext) {
			pairDevice(ctx, "tv-1")
			Expect(bind(ctx, "tv-1", "table")).To(Succeed())

			room, err := service.Room(authoredRoom{devices.NewRoom("room", "")}).CurrentRoom(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(room.DevicesFor("table", devices.KindScreen)).To(ConsistOf(And(
				HaveField("ID", "tv-1"),
				HaveField("Name", "Table TV"),
				HaveField("Integration", devices.IntegrationScreen),
			)))
		}, SpecTimeout(15*time.Second))

		It("removes the binding with an empty target", func(ctx SpecContext) {
			pairDevice(ctx, "tv-1")
			Expect(bind(ctx, "tv-1", "table")).To(Succeed())
			Expect(bind(ctx, "tv-1", "")).To(Succeed())

			room, err := service.Room(authoredRoom{devices.NewRoom("room", "")}).CurrentRoom(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(room.DevicesFor("table", devices.KindScreen)).To(BeEmpty())
		}, SpecTimeout(15*time.Second))

		It("leaves the authored room as it is, and keeps its bindings", func(ctx SpecContext) {
			authored := devices.NewRoom("room", "Living room")
			authored.AddDevice(&devices.Device{ID: "wall", Kind: devices.KindScreen, Integration: devices.IntegrationScreen})
			Expect(authored.Bind("table", "wall")).To(Succeed())

			pairDevice(ctx, "tv-1")
			Expect(bind(ctx, "tv-1", "table")).To(Succeed())

			room, err := service.Room(authoredRoom{authored}).CurrentRoom(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(room.Name).To(Equal("Living room"))
			Expect(room.Bindings["table"]).To(ConsistOf("wall", "tv-1"))

			Expect(authored.Bindings["table"]).To(ConsistOf("wall"))
			Expect(authored.Devices).To(HaveLen(1))
		}, SpecTimeout(15*time.Second))

		It("forgets the binding of a revoked device", func(ctx SpecContext) {
			pairDevice(ctx, "tv-1")
			Expect(bind(ctx, "tv-1", "table")).To(Succeed())
			_, err := client.Revoke(withAuthorization(ctx, admin), connect.NewRequest(&devicesv1.RevokeDeviceRequest{Id: "tv-1"}))
			Expect(err).ShouldNot(HaveOccurred())

			room, err := service.Room(authoredRoom{devices.NewRoom("room", "")}).CurrentRoom(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(room.Bindings).To(BeEmpty())
		}, SpecTimeout(15*time.Second))

		It("answers NotFound for a device that isn't paired", func(ctx SpecContext) {
			Expect(bind(ctx, "tv-1", "table")).To(HaveConnectCode(connect.CodeNotFound))
		})

		It("refuses an unusable target", func(ctx SpecContext) {
			pairDevice(ctx, "tv-1")
			Expect(bind(ctx, "tv-1", "../table")).To(HaveConnectCode(connect.CodeInvalidArgument))
		}, SpecTimeout(15*time.Second))
	})
})

// authoredRoom is a room as it comes from file storage.
type authoredRoom struct {
	room *devices.Room
}

func (a authoredRoom) CurrentRoom(context.Context) (*devices.Room, error) {
	return a.room, nil
}

type authorization_key struct{}

// withAuthorization makes unary calls with ctx send the Authorization header, see [authorizationFromContext].
func withAuthorization(ctx context.Context, authorization string) context.Context {
	return context.WithValue(ctx, authorization_key{}, authorization)
}

func authorizationFromContext() connect.ClientOption {
	return connect.WithInterceptors(connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if authorization, ok := ctx.Value(authorization_key{}).(string); ok {
				req.Header().Set("Authorization", authorization)
			}

			return next(ctx, req)
		}
	}))
}
