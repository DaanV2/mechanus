package persistence

import (
	"context"
	"fmt"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xgorm"
	"github.com/charmbracelet/log"
)

// ApplyMigrations applies database migrations
func ApplyMigrations(ctx context.Context, db *DB, models ...any) error {
	log.Debug("Applying database migrations")

	// TODO: Implement more sophisticated migration logic if needed
	// This is a simple auto-migration approach
	ctx = xgorm.WithPrefix(ctx, "migration")
	err := db.gormDB.WithContext(ctx).AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
