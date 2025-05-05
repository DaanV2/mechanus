package database

import "fmt"

// ApplyMigrations applies database migrations
func ApplyMigrations(db *DB, models ...interface{}) error {
	// TODO: Implement more sophisticated migration logic if needed
	// This is a simple auto-migration approach
	err := db.gormDB.AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}