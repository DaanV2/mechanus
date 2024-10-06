package database_test

import (
	"testing"

	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/stretchr/testify/require"
)

type TestItem struct {
	database.BaseItem
	Name string
}

func Test_Database(t *testing.T) {
	mem := database.NewMemoryIO()
	db := database.NewDatabaseWith(mem)

	for range 100 {
		item := TestItem{
			BaseItem: database.NewBaseItem(),
			Name:     "gandalf",
		}

		table := database.GetTable[TestItem](db, "test")

		temp, err := table.Get(item.ID)
		require.Error(t, err)
		require.Empty(t, temp)

		err = table.Set(item.ID, item)
		require.NoError(t, err)

		temp, err = table.Get(item.ID)
		require.NoError(t, err)

		// Time is weird
		temp.CreatedAt = item.CreatedAt
		temp.UpdatedAt = item.UpdatedAt

		require.Equal(t, temp, item)
	}
}
