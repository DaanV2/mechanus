package main

import (
	"errors"
	"log"

	"github.com/DaanV2/mechanus/server/cmd"
	"github.com/DaanV2/mechanus/server/internal/setup"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	setup.Viper()
	setup.Logger()
	setup.Folders()
	setup.Config()

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
