package components

import (
	"context"

	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/database/models"
	"github.com/charmbracelet/log"
)

func GetDatabaseOptions() ([]database.Option, error) {
	return database.GetOptions()
}

func SetupTestDatabase(dbOptions ...database.Option) (*database.DB, error) {
	dbOptions = append(dbOptions, database.WithType(database.InMemory))

	return setupDatabase(context.Background(), dbOptions...)
}

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
