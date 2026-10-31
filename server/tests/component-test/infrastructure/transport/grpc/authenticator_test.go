package grpc_test

import (
	"context"
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
		// The context the next handler in the chain received
		received context.Context
	)

	BeforeEach(func(setupCtx SpecContext) {
		db := util_test.CreateDatabase(setupCtx)
		jtiService = authentication.NewJTIService(db)
		keyManager, err := authentication.NewKeyManager(storage.DBStorage[*authentication.KeyData](db))
		Expect(err).ShouldNot(HaveOccurred())
		jwtService = authentication.NewJWTService(jtiService, keyManager)

		received = nil
		handler = grpc.NewAuthenticator(jwtService).WrapUnary(func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
			received = ctx

			return nil, nil
		})
	})

	call := func(ctx context.Context, authorization string) {
		req := connect.NewRequest(&usersv1.ListUsersRequest{})
		if authorization != "" {
			req.Header().Set("Authorization", authorization)
		}

		Expect(func() { _, _ = handler(ctx, req) }).ToNot(Panic())
		Expect(received).ToNot(BeNil())
	}

	createToken := func(ctx context.Context, scope string) string {
		token, err := jwtService.Create(ctx, util_test.CreateUser(), authentication.GrantPassword, scope)
		Expect(err).ShouldNot(HaveOccurred())

		return token
	}

	It("leaves the request anonymous without a token", func(ctx SpecContext) {
		call(ctx, "")

		_, err := authentication.JWTFromContext(received)
		Expect(err).To(MatchError(authentication.ErrNoToken))
	})

	It("marks a valid token as authenticated", func(ctx SpecContext) {
		call(ctx, "Bearer "+createToken(ctx, authentication.ScopeAll))

		jwt, err := authentication.JWTFromContext(received)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(jwt.IsAuthenicated()).To(BeTrue())
	})

	DescribeTable("treats an unparsable token as anonymous",
		func(ctx SpecContext, authorization string) {
			call(ctx, authorization)

			Expect(authentication.IsAuthenicated(received)).To(BeFalse())
		},
		Entry("malformed", "Bearer not.a.jwt"),
		Entry("garbage", "Bearer garbage"),
		Entry("empty bearer", "Bearer "),
	)

	It("treats an expired token as anonymous", func(ctx SpecContext) {
		jwtService.Options().TokenDuration = -time.Hour
		token := createToken(ctx, authentication.ScopeAll)

		call(ctx, "Bearer "+token)

		Expect(authentication.IsAuthenicated(received)).To(BeFalse())
	})

	It("treats a tampered token as anonymous", func(ctx SpecContext) {
		token := createToken(ctx, authentication.ScopeAll)
		// Flip the last character of the signature
		last := token[len(token)-1]
		replacement := byte('A')
		if last == 'A' {
			replacement = 'B'
		}
		token = token[:len(token)-1] + string(replacement)

		call(ctx, "Bearer "+token)

		Expect(authentication.IsAuthenicated(received)).To(BeFalse())
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

		Expect(authentication.IsAuthenicated(received)).To(BeFalse())
	})

	It("does not authenticate a refresh-only token", func(ctx SpecContext) {
		call(ctx, "Bearer "+createToken(ctx, authentication.ScopeRefresh))

		jwt, err := authentication.JWTFromContext(received)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(jwt.IsAuthenicated()).To(BeFalse())
	})
})
