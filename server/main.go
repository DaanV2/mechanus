/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"errors"
	"log"

	"github.com/DaanV2/mechanus/server/cmd"
	"github.com/DaanV2/mechanus/server/internal/setup"
	"github.com/DaanV2/mechanus/server/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	cobra.OnInitialize(
		setup.Logger,
		setup.Config,
		func() {
			logconfing := config.Get[config.Logger]()
			setup.UpdateLogger(
				logconfing.ReportCaller,
				logconfing.Level,
				logconfing.Format,
			)
		})

	cobra.OnFinalize(func() {
		err := viper.SafeWriteConfig()
		var verr viper.ConfigFileAlreadyExistsError
		if errors.As(err, &verr) {
			return
		}

		log.Fatal("troubling saving config", "error", err)
	})

	cmd.Execute()
}
