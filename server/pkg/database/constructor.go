package database

import (
	"context"
	"fmt"
	"time"

	xgorm "github.com/DaanV2/mechanus/server/pkg/extensions/gorm"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	gormDB *gorm.DB
}

func (db *DB) Close() error {
	sqlDB, err := db.gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	return sqlDB.Close()
}

func (db *DB) WithContext(ctx context.Context) *gorm.DB {
	return db.gormDB.WithContext(ctx)
}

// NewDB creates a new database connection with the given options
func NewDB(opts ...Option) (*DB, error) {
	// Default configuration
	config := &Config{
		Type:            SQLite,
		DSN:             "db.sqlite",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// GORM configuration
	gormConfig := &gorm.Config{
		Logger: xgorm.NewGormlogger().LogMode(config.LogLevel),
	}

	if config.Logger != nil {
		gormConfig.Logger = config.Logger
	}

	var dailer gorm.Dialector

	// Connect to the database based on the type
	switch config.Type {
	case SQLite:
		dailer = sqlite.Open(config.DSN)
	case InMemory:
		dailer = sqlite.Open("file::memory:")
	case PostgreSQL:
		dailer = postgres.Open(config.DSN)
	case MySQL:
		dailer = mysql.Open(config.DSN)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	db, err := gorm.Open(dailer, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	return &DB{gormDB: db}, nil
}
