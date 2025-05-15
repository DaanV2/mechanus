package setup

import (
	"errors"

	"github.com/DaanV2/mechanus/server/pkg/config"
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

// Config loads the necassary files, and values
func Config() {
	if err := viper.ReadInConfig(); err != nil {
		var verr viper.ConfigFileNotFoundError
		if errors.As(err, &verr) {
			log.Debug("couldn't find a log file, falling back to defaults, arguments and environment files", "error", err)
			// Config file not found; ignore error if desired
		} else {
			log.Fatal("error during reading config file", "error", err)
		}
	}
}

func Viper() {
	viper.SetEnvKeyReplacer(config.EnvironmentNamer())
	viper.AutomaticEnv()
	viper.AddConfigPath(config.UserConfigDir())
	viper.SetConfigType("yaml")
}
