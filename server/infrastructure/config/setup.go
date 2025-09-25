package config

import (
	"errors"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

// Config loads the necassary files, and values
func SetupConfig() {
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		var verr viper.ConfigFileNotFoundError
		if errors.As(err, &verr) {
			log.Debug("couldn't find a log file, falling back to defaults, arguments and environment files", "error", err)
			// Config file not found; ignore error if desired

			return
		}

		log.Fatal("error during reading config file", "error", err)
	}
}

func SetupViper() {
	viper.SetEnvKeyReplacer(EnvironmentNamer())
	viper.SetConfigType("yaml")

	for _, v := range ConfigPaths() {
		viper.AddConfigPath(v)
	}
}
