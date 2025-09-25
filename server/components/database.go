package components

import (
	"context"

	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/charmbracelet/log"
)

// GetOptions builds a list of options from flags/envs variable via viper.
// See [persistence.GetOptions] for available values
func GetDatabaseOptions() ([]persistence.Option, error) {
	return persistence.GetOptions()
}

// SetupTestDatabase setups a database that is stored inmemory and doesn't write to files.
func SetupTestDatabase(setupCtx context.Context, dbOptions ...persistence.Option) (*persistence.DB, error) {
	dbOptions = append(dbOptions, persistence.WithType(persistence.InMemory))

	return setupDatabase(setupCtx, dbOptions...)
}

// SetupDatabase returns a new [persistence.DB] configured with the given options.
// Use [GetDatabaseOptions] to get a set of base options
func SetupDatabase(setupCtx context.Context, dbOptions ...persistence.Option) (*persistence.DB, error) {
	return setupDatabase(setupCtx, dbOptions...)
}

func setupDatabase(setupCtx context.Context, dbOptions ...persistence.Option) (*persistence.DB, error) {
	log.Debug("Setting up a database")

	db, err := persistence.NewDB(dbOptions...)
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

	if err := persistence.ApplyMigrations(setupCtx, db, m...); err != nil {
		return nil, err
	}

	return db, nil
}
