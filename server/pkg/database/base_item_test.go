package database_test

import (
	"testing"
	"time"

	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/stretchr/testify/require"
)

func Test_BaseItem(t *testing.T) {
	t.Run("New created base should have valid uuid and dates", func(t *testing.T) {
		n := time.Now()
		b := database.NewBaseItem()

		require.Len(t, b.ID, 36)
		require.GreaterOrEqual(t, b.CreatedAt, n)
		require.GreaterOrEqual(t, b.UpdatedAt, n)
		require.Nil(t, b.DeletedAt)
		require.False(t, b.IsDeleted())
	})

	t.Run("If updated, then the updated field should be updated", func(t *testing.T) {
		b := database.NewBaseItem()

		time.Sleep(time.Second)

		newb := b.Update()
		require.Greater(t, newb.UpdatedAt, b.UpdatedAt)
		require.Equal(t, newb.CreatedAt, b.CreatedAt)
		require.Equal(t, newb.DeletedAt, b.DeletedAt)
	})

	t.Run("If deleted, then the deleted field should be updated", func(t *testing.T) {
		b := database.NewBaseItem()

		time.Sleep(time.Second)

		newb := b.Delete()
		require.Equal(t, newb.CreatedAt, b.CreatedAt)
		require.Greater(t, newb.UpdatedAt, b.UpdatedAt)
		require.NotNil(t, newb.DeletedAt)
		require.True(t, newb.IsDeleted())
	})
}