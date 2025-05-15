package database

import (
	"fmt"

	"github.com/charmbracelet/log"
)

// ApplyMigrations applies database migrations
func ApplyMigrations(db *DB, models ...any) error {
	log.Debug("Applying database migrations")

	// TODO: Implement more sophisticated migration logic if needed
	// This is a simple auto-migration approach
	err := db.gormDB.AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}
