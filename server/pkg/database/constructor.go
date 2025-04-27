package database

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB wraps the gorm.DB instance and provides a Close method
type DB struct {
	gormDB *gorm.DB
}

func (d *DB) Close() error {
	sqlDB, err := d.gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}
	return sqlDB.Close()
}


// NewDB creates a new database connection with the given options
func NewDB(opts ...Option) (*DB, error) {
	// Default configuration
	config := &Config{
		Type:           SQLite,
		DSN:            "db.sqlite",
		MaxIdleConns:   10,
		MaxOpenConns:   100,
		ConnMaxLifetime: time.Hour,
		SlowThreshold:  time.Second,
		LogWriter:      log.Default().WithPrefix("db").StandardLog(),
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Configure GORM logger
	gormLogger := logger.New(
		config.LogWriter,
		logger.Config{
			SlowThreshold:             config.SlowThreshold,
			LogLevel:                  config.LogLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// GORM configuration
	gormConfig := &gorm.Config{
		Logger: gormLogger,
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

	return &DB{db}, nil
}