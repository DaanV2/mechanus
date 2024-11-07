package config

import (
	"path/filepath"

	"github.com/spf13/pflag"
)

type (
	DatabaseConfig struct {
		Mode Flag[string]
	}

	DatabaseIOConfig struct {
		Folder string
	}
)

var (
	Database = &DatabaseConfig{
		Mode: String("database-mode", "files", "The mode type of the database: acceptable modes: files"),
	}

	DatabaseIO = &DatabaseIOConfig{
		Folder: filepath.Join(UserCacheDir(), "db"),
	}
)

func (c *DatabaseConfig) AddToSet(set *pflag.FlagSet) {

}
