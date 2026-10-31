package grpc_test

import (
	"context"
	"strings"
	"time"

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"
	usersv1 "github.com/DaanV2/mechanus/server/proto/users/v1"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
)

var _ = Describe("Authenticator", func() {
	var (
		jtiService *authentication.JTIService
		jwtService *authentication.JWTService
		handler    connect.UnaryFunc

		// What the next handler in the chain found on its context
		called bool
		jwt    authentication.JWTContext
		jwtErr error
	)

	BeforeEach(func(setupCtx SpecContext) {
		db := util_test.CreateDatabase(setupCtx)
		jtiService = authentication.NewJTIService(db)
		keyManager, err := authentication.NewKeyManager(storage.DBStorage[*authentication.KeyData](db))
		Expect(err).ShouldNot(HaveOccurred())
		jwtService = authentication.NewJWTService(jtiService, keyManager)

		called = false
		handler = grpc.NewAuthenticator(jwtService).WrapUnary(func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
			called = true
			jwt, jwtErr = authentication.JWTFromContext(ctx)

			return connect.NewResponse(&usersv1.ListUsersResponse{}), nil
		})
	})

	call := func(ctx context.Context, authorization string) {
		req := connect.NewRequest(&usersv1.ListUsersRequest{})
		if authorization != "" {
			req.Header().Set("Authorization", authorization)
		}

		Expect(func() { _, _ = handler(ctx, req) }).ToNot(Panic())
		Expect(called).To(BeTrue())
	}

	createToken := func(ctx context.Context, scope string) string {
		token, err := jwtService.Create(ctx, util_test.CreateUser(), authentication.GrantPassword, scope)
		Expect(err).ShouldNot(HaveOccurred())

		return token
	}

	It("leaves the request anonymous without a token", func(ctx SpecContext) {
		call(ctx, "")

		Expect(jwtErr).To(MatchError(authentication.ErrNoToken))
	})

	It("marks a valid token as authenticated", func(ctx SpecContext) {
		call(ctx, "Bearer "+createToken(ctx, authentication.ScopeAll))

		Expect(jwtErr).ShouldNot(HaveOccurred())
		Expect(jwt.IsAuthenicated()).To(BeTrue())
	})

	DescribeTable("treats an unparsable token as anonymous",
		func(ctx SpecContext, authorization string) {
			call(ctx, authorization)

			Expect(jwt.IsAuthenicated()).To(BeFalse())
		},
		Entry("malformed", "Bearer not.a.jwt"),
		Entry("garbage", "Bearer garbage"),
		Entry("empty bearer", "Bearer "),
	)

	It("treats an expired token as anonymous", func(ctx SpecContext) {
		jwtService.Options().TokenDuration = -time.Hour
		token := createToken(ctx, authentication.ScopeAll)

		call(ctx, "Bearer "+token)

		Expect(jwt.IsAuthenicated()).To(BeFalse())
	})

	It("treats a tampered token as anonymous", func(ctx SpecContext) {
		token := createToken(ctx, authentication.ScopeAll)
		// Change a character in the middle of the signature. Not the last one: its low bits are base64 padding,
		// so changing it can decode to the same signature bytes.
		i := strings.LastIndex(token, ".") + 10
		replacement := byte('A')
		if token[i] == 'A' {
			replacement = 'B'
		}
		token = token[:i] + string(replacement) + token[i+1:]

		call(ctx, "Bearer "+token)

		Expect(jwt.IsAuthenicated()).To(BeFalse())
	})

	It("marks a token with a revoked JTI as not authenticated", func(ctx SpecContext) {
		token := createToken(ctx, authentication.ScopeAll)
		t, err := jwtService.Validate(ctx, token)
		Expect(err).ShouldNot(HaveOccurred())
		claims, ok := authentication.GetClaims(t.Claims)
		Expect(ok).To(BeTrue())
		_, err = jtiService.Revoke(ctx, claims.ID)
		Expect(err).ShouldNot(HaveOccurred())

		call(ctx, "Bearer "+token)

		Expect(jwtErr).ShouldNot(HaveOccurred())
		Expect(jwt.IsAuthenicated()).To(BeFalse())
	})

	It("does not authenticate a refresh-only token", func(ctx SpecContext) {
		call(ctx, "Bearer "+createToken(ctx, authentication.ScopeRefresh))

		Expect(jwtErr).ShouldNot(HaveOccurred())
		Expect(jwt.IsAuthenicated()).To(BeFalse())
	})
})
