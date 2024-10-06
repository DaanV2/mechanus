package config

import "github.com/spf13/pflag"

type DatabaseConfig struct {
	Folder Flag[string]
}

var Database = &DatabaseConfig{
	Folder: String("database.folder", "./db", "The folder to store database files in"),
}

func (c *DatabaseConfig) AddToSet(set *pflag.FlagSet) {
	c.Folder.AddToSet(set)
}
