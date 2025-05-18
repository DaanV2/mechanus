package components

import (
	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/database/models"
	"github.com/charmbracelet/log"
)

func SetupTestDatabase(dbOptions ...database.Option) (*database.DB, error) {
	dbOptions = append(dbOptions, database.WithType(database.InMemory) )

	return setupDatabase(dbOptions...)
}

func SetupDatabase(dbOptions ...database.Option) (*database.DB, error) {
	opts := []database.Option{
		database.WithType(database.SQLite),
	}

	opts = append(opts, dbOptions...)

	return setupDatabase(opts...)
}

func setupDatabase(dbOptions ...database.Option) (*database.DB, error) {
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

	if err := database.ApplyMigrations(db, m...); err != nil {
		return nil, err
	}

	return db, nil
}
