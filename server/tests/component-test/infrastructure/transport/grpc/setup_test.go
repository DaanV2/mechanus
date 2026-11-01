package grpc_test

import (
	"context"

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
)

// authEnv holds the services needed to test the authenticated gRPC handlers against a test database.
type authEnv struct {
	db    *persistence.DB
	users *application.UserService
	jtis  *authentication.JTIService
	jwts  *authentication.JWTService
}

func newAuthEnv(ctx context.Context) *authEnv {
	GinkgoHelper()

	db := util_test.CreateDatabase(ctx)
	keys, err := authentication.NewKeyManager(storage.DBStorage[*authentication.KeyData](db))
	Expect(err).ShouldNot(HaveOccurred())
	jtis := authentication.NewJTIService(db)

	return &authEnv{
		db:    db,
		users: application.NewUserService(repositories.NewUserRepository(db)),
		jtis:  jtis,
		jwts:  authentication.NewJWTService(jtis, keys),
	}
}

// createUser stores a new user with the given password and roles.
func (e *authEnv) createUser(ctx context.Context, password string, roles ...string) *models.User {
	GinkgoHelper()

	user := &models.User{
		Username:     "user_" + util_test.CreateUserID(),
		PasswordHash: []byte(password),
		Roles:        roles,
	}
	Expect(e.users.Create(ctx, user)).To(Succeed())

	return user
}

// loginAs starts a session for the user, and returns a context authenticated as that session, like the Authenticator would.
func (e *authEnv) loginAs(ctx context.Context, user *models.User) (context.Context, *authentication.JWTClaims) {
	GinkgoHelper()

	tokens, err := e.jwts.CreateSession(ctx, user, authentication.GrantPassword)
	Expect(err).ShouldNot(HaveOccurred())
	claims := e.claims(ctx, tokens.Access)

	// Detached from ctx, so contexts made in a BeforeEach outlive its SpecContext
	return authentication.ContextWithJWT(context.WithoutCancel(ctx), claims, true), claims
}

func (e *authEnv) claims(ctx context.Context, token string) *authentication.JWTClaims {
	GinkgoHelper()

	t, err := e.jwts.Validate(ctx, token)
	Expect(err).ShouldNot(HaveOccurred())
	claims, ok := authentication.GetClaims(t.Claims)
	Expect(ok).To(BeTrue())

	return claims
}

// isRevoked returns whenever or not the session (JTI) is revoked.
func (e *authEnv) isRevoked(ctx context.Context, jti string) bool {
	GinkgoHelper()

	result, err := e.jtis.Get(ctx, jti)
	Expect(err).ShouldNot(HaveOccurred())

	return result.Revoked
}

// HaveConnectCode matches a connect error with the given code.
func HaveConnectCode(code connect.Code) OmegaMatcher {
	return WithTransform(connect.CodeOf, Equal(code))
}
