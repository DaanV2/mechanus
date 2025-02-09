package authenication_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/DaanV2/mechanus/server/pkg/models"
	"github.com/DaanV2/mechanus/server/pkg/models/roles"
	"github.com/DaanV2/mechanus/server/pkg/models/users"
	memory_storage "github.com/DaanV2/mechanus/server/pkg/storage/memory"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func Test_JWT_Service(t *testing.T) {
	jtiStorage := memory_storage.NewStorage[[]authenication.JTI]()
	keyStorage := memory_storage.NewStorage[*authenication.KeyData]()
	keyManager, err := authenication.NewKeyManager(keyStorage)
	require.NoError(t, err)
	jtiService := authenication.NewJTIService(jtiStorage)

	jwtService := authenication.NewJWTService(jtiService, keyManager)
	require.NoError(t, err)

	user := users.User{
		BaseItem:     models.NewBaseItem(),
		Username:     "gandalf",
		Roles:        []roles.Role{roles.ADMIN, roles.OPERATOR},
		Campaigns:    []string{"lord-of-the-rings"},
		PasswordHash: []byte("the-eagles"),
	}

	t.Run("can create a jwt", func(t *testing.T) {
		jwt, err := jwtService.Create(context.Background(), user, "password")
		require.NoError(t, err)
		t.Log("jwt", jwt)
	})

	t.Run("can verify its own jwt", func(t *testing.T) {
		jwt, err := jwtService.Create(context.Background(), user, "password")
		require.NoError(t, err)
		t.Log("jwt", jwt)

		token, err := jwtService.Validate(context.Background(), jwt)
		require.NoError(t, err)

		c, ok := authenication.GetClaims(token.Claims)
		require.True(t, ok)

		require.Equal(t, c.User.ID, user.ID)
		require.Equal(t, c.User.Name, user.Username)
		require.Equal(t, c.User.Roles, user.Roles)
		require.Equal(t, c.User.Campaigns, user.Campaigns)
	})
}

func Test_JWT_Claims(t *testing.T) {
	original := authenication.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    authenication.JWT_ISSUER,
			Audience:  []string{authenication.JWT_AUDIENCE},
			Subject:   "movie",
			ExpiresAt: nil,
			NotBefore: nil,
			IssuedAt:  nil,
			ID:        "0123456789",
		},
		User: authenication.JWTUser{
			ID:        "0123456789",
			Name:      "gandalf",
			Roles:     []roles.Role{roles.ADMIN, roles.OPERATOR},
			Campaigns: []string{"lord-of-the-rings"},
		},
		Scope: "password",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var claim authenication.JWTClaims
	err = json.Unmarshal([]byte(data), &claim)
	require.NoError(t, err)
	require.Equal(t, claim, original)
}
