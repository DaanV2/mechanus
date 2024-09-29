package setup

import (
	"errors"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

// Config loads the necassary files, and values
func Config() {
	viper.AutomaticEnv()
	viper.AddConfigPath(".")
	viper.AddConfigPath(".config/")
	viper.AddConfigPath("/etc/mechanus/")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		var verr viper.ConfigFileNotFoundError
		if errors.As(err, &verr) {
			log.Debug("couldn't find a log file, falling back to defaults, arguments and enviroment files", "error", err)
			// Config file not found; ignore error if desired
		} else {
			log.Fatal("error during reading config file", "error", err)
		}
	}
}