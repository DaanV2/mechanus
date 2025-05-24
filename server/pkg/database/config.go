package database

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/DaanV2/mechanus/server/mechanus/paths"
	"github.com/DaanV2/mechanus/server/pkg/config"
	"gorm.io/gorm/logger"
)

// DBType represents the type of database to use
type DBType string

func (db DBType) String() string {
	return string(db)
}

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

var (
	DatabaseConfig      = config.New("database").WithValidate(validateDatabaseFlags)
	TypeFlag            = DatabaseConfig.String("database.type", SQLite.String(), "The type of database to connect/use: supported values: sqlite, postgres, mysql. (For testing purposes there is also inmemory)")
	DSNFlag             = DatabaseConfig.String("database.dsn", "db.sqlite", "A datasource name, depends on type of database, but usually referes to file name or the connection string")
	MaxIdleConnsFlag    = DatabaseConfig.Int("database.maxIdleConns", 2, "Sets the maximum number of connections in the idle connection pool. If n <= 0, no idle connections are retained.")
	MaxOpenConnsFlag    = DatabaseConfig.Int("database.maxOpenConns", 0, "Sets the maximum number of open connections to the database. If n <= 0, then there is no limit on the number of open connections.")
	ConnMaxLifetimeFlag = DatabaseConfig.Duration("database.connMaxLifetime", 1*time.Hour, "Sets the maximum amount of time a connection may be reused. If d <= 0, connections are not closed due to a connection's age.")
)

func validateDatabaseFlags(conf *config.Config) error {
	var err error

	dbt := conf.GetString("database.type")
	switch DBType(dbt) {
	case MySQL, SQLite, InMemory, PostgreSQL:
	default:
		err = errors.Join(err, errors.New("unknown database type: "+dbt))
	}

	return err
}

func GetOptions() ([]Option, error) {
	opts := []Option{
		WithMaxIdleConns(DatabaseConfig.GetInt("database.maxIdleConns")),
		WithMaxOpenConns(DatabaseConfig.GetInt("database.maxOpenConns")),
		WithConnMaxLifetime(DatabaseConfig.GetDuration("database.connMaxLifetime")),
	}

	dt := DBType(DatabaseConfig.GetString("database.type"))
	dsn := DatabaseConfig.GetString("database.dsn")

	// If SQLITE and dsn is empty or default, we will sanitize to the state directory
	if dt == SQLite && (dsn == "" || dsn == "db.sqlite") {
		d, err := paths.GetStateDir()
		if err != nil {
			return nil, fmt.Errorf("cannot seem to determine the state directory: %w", err)
		}

		dsn = filepath.Join(d, "db.sqlite")
	}

	opts = append(opts, WithType(dt), WithDSN(dsn))

	return opts, nil
}
