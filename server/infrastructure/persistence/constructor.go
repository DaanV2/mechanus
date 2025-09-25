package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xgorm"
	"github.com/charmbracelet/log"
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
		MaxIdleConns:    2,
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

	// Connect to the database based on the type
	log.WithPrefix("db").Debug("opening database", "type", config.Type, "dsn", config.DSN)
	var dailer gorm.Dialector
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

	log.WithPrefix("db").Debug("applying database settings", "max idle conns", config.MaxIdleConns, "max open conns", config.MaxOpenConns, "conn max lifetime", config.ConnMaxLifetime)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	return &DB{gormDB: db}, nil
}
