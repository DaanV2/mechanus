// The main entry point of any mechanus program, setups config, logging and folders then goes to ./cmd
package main

import (
	"errors"

	"github.com/DaanV2/mechanus/server/cmd"
	"github.com/DaanV2/mechanus/server/infrastructure/config"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	config.SetupViper()
	logging.SetupLogger()
	storage.SetupFolders()
	config.SetupConfig()

	cobra.OnFinalize(func() {
		err := viper.SafeWriteConfig()

		var verr viper.ConfigFileAlreadyExistsError
		if err == nil || errors.As(err, &verr) {
			return
		}

		log.Fatal("troubling saving config", "error", err)
	})

	cmd.Execute()
}
