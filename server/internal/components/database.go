package components

import (
	"context"

	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/database/models"
	"github.com/charmbracelet/log"
)

// GetOptions builds a list of options from flags/envs variable via viper.
// See [database.GetOptions] for available values
func GetDatabaseOptions() ([]database.Option, error) {
	return database.GetOptions()
}

// SetupTestDatabase setups a database that is stored inmemory and doesn't write to files.
func SetupTestDatabase(dbOptions ...database.Option) (*database.DB, error) {
	dbOptions = append(dbOptions, database.WithType(database.InMemory))

	return setupDatabase(context.Background(), dbOptions...)
}

// SetupDatabase returns a new [database.DB] configured with the given options.
// Use [GetDatabaseOptions] to get a set of base options
func SetupDatabase(dbOptions ...database.Option) (*database.DB, error) {
	return setupDatabase(context.Background(), dbOptions...)
}

func setupDatabase(ctx context.Context, dbOptions ...database.Option) (*database.DB, error) {
	log.Debug("Setting up a database")

	db, err := database.NewDB(dbOptions...)
	if err != nil {
		return nil, err
	}

	m := []any{
		&models.User{},
		&models.Campaign{},
		&models.Character{},
		&models.JTI{},
		&models.KeyValue{},
	}

	if err := database.ApplyMigrations(ctx, db, m...); err != nil {
		return nil, err
	}

	return db, nil
}
