package database

import (
	"time"

	"gorm.io/gorm/logger"
)

// DBType represents the type of database to use
type DBType string

const (
	// SQLite database type
	SQLite DBType = "sqlite"
	// InMemory SQLite database
	InMemory DBType = "memory"
	// PostgreSQL database type
	PostgreSQL DBType = "postgres"
	// MySQL database type
	MySQL DBType = "mysql"
)

// Config holds the database configuration
type Config struct {
	Type            DBType
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	LogLevel        logger.LogLevel
	Logger          logger.Interface
}
