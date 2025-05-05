package setup

import (
	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/database/models"
	"github.com/charmbracelet/log"
)

func SetupDatabase(dbOptions ...database.Option) (*database.DB, error) {
	log.Debug("Setting up a database")

	db, err := database.NewDB(dbOptions...)
	if err != nil {
		return nil, err
	}

	models := []any{
		&models.User{},
		&models.Campaign{},
		&models.Character{},
	}

	if err := database.ApplyMigrations(db, models...); err != nil {
		return nil, err
	}

	return db, nil
}