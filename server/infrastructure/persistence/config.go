package persistence

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/config"
	"github.com/DaanV2/mechanus/server/pkg/paths"
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
type DatabaseConfig struct {
	Type            DBType
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	LogLevel        logger.LogLevel
	Logger          logger.Interface
}

var (
	DatabaseConfigSet   = config.New("database").WithValidate(validateDatabaseFlags)
	TypeFlag            = DatabaseConfigSet.String("database.type", SQLite.String(), "The type of database to connect/use: supported values: sqlite, postgres, mysql. (For testing purposes there is also inmemory)")
	DSNFlag             = DatabaseConfigSet.String("database.dsn", "db.sqlite", "A datasource name, depends on type of database, but usually referes to file name or the connection string")
	MaxIdleConnsFlag    = DatabaseConfigSet.Int("database.max-idle-conss", 2, "Sets the maximum number of connections in the idle connection pool. If n <= 0, no idle connections are retained.")
	MaxOpenConnsFlag    = DatabaseConfigSet.Int("database.max-open-conns", 0, "Sets the maximum number of open connections to the database. If n <= 0, then there is no limit on the number of open connections.")
	ConnMaxLifetimeFlag = DatabaseConfigSet.Duration("database.conn-max-lifetime", 1*time.Hour, "Sets the maximum amount of time a connection may be reused. If d <= 0, connections are not closed due to a connection's age.")
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

// GetOptions builds a list of options from flags/envs variable via viper.
// See [DatabaseConfig] for available values
func GetOptions() ([]Option, error) {
	opts := []Option{
		WithMaxIdleConns(MaxIdleConnsFlag.Value()),
		WithMaxOpenConns(MaxOpenConnsFlag.Value()),
		WithConnMaxLifetime(ConnMaxLifetimeFlag.Value()),
	}

	dt := DBType(TypeFlag.Value())
	dsn := DSNFlag.Value()

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
