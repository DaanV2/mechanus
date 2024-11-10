package users_test

import (
	"testing"
	"time"

	"github.com/DaanV2/mechanus/server/pkg/models"
	memory_storage "github.com/DaanV2/mechanus/server/pkg/storage/memory"
	"github.com/DaanV2/mechanus/server/services/users"
	"github.com/stretchr/testify/require"
)

func Test_User_Service(t *testing.T) {
	original := models.User{
		Name:         "gandalf",
		Roles:        []string{"admin"},
		Campaigns:    []string{"homebrew"},
		PasswordHash: []byte("good-old-password"),
	}

	t.Run("Can create a new user", func(t *testing.T) {
		store := memory_storage.NewStorage[models.User]()
		now := time.Now()
		service := users.NewService(store)

		// Create
		u, err := service.Create(original)
		require.NoError(t, err)

		// Validate
		require.NotEqual(t, u.PasswordHash, original.PasswordHash)
		require.Equal(t, u.Name, original.Name)
		require.Equal(t, u.Roles, original.Roles)
		require.Equal(t, u.Campaigns, original.Campaigns)

		require.Len(t, u.ID, 36)
		require.GreaterOrEqual(t, u.CreatedAt, now)
		require.GreaterOrEqual(t, u.UpdatedAt, now)
		require.Nil(t, u.DeletedAt)
		require.False(t, u.IsDeleted())

		getUser, err := service.Get(u.ID)
		require.NoError(t, err)

		// Time comparing is weird on deep equal
		require.WithinDuration(t, getUser.CreatedAt, u.CreatedAt, time.Second)
		require.WithinDuration(t, getUser.UpdatedAt, u.UpdatedAt, time.Second)

		getUser.CreatedAt = u.CreatedAt
		getUser.UpdatedAt = u.UpdatedAt
		require.Equal(t, getUser, u)
	})
}
