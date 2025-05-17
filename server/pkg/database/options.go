package database

import (
	"time"

	"gorm.io/gorm/logger"
)

// Option is a function that modifies the Config
type Option func(*Config)

// WithType sets the database type
func WithType(dbType DBType) Option {
	return func(c *Config) {
		c.Type = dbType
	}
}

// WithDSN sets the data source name
func WithDSN(dsn string) Option {
	return func(c *Config) {
		c.DSN = dsn
	}
}

// WithMaxIdleConns sets the maximum number of idle connections
func WithMaxIdleConns(n int) Option {
	return func(c *Config) {
		c.MaxIdleConns = n
	}
}

// WithMaxOpenConns sets the maximum number of open connections
func WithMaxOpenConns(n int) Option {
	return func(c *Config) {
		c.MaxOpenConns = n
	}
}

// WithConnMaxLifetime sets the maximum lifetime of a connection
func WithConnMaxLifetime(d time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxLifetime = d
	}
}

// WithDBLogLevel sets the log level for the database
func WithDBLogLevel(level logger.LogLevel) Option {
	return func(c *Config) {
		c.LogLevel = level
	}
}

// WithDBLogger provides the db with a new logger, ignores [WithDBLogLevel]
func WithDBLogger(logger logger.Interface) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}