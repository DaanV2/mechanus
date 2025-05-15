package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DaanV2/mechanus/server/internal/components"
	"github.com/DaanV2/mechanus/server/internal/grpc"
	"github.com/DaanV2/mechanus/server/internal/web"
	"github.com/DaanV2/mechanus/server/pkg/application"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	RunE:  ServerWorkload,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Info("server starting")
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		log.Info("server stopped")
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// flags := serverCmd.Flags()
	web.WebConfig.AddToSet(serverCmd.Flags())
	grpc.APIConfig.AddToSet(serverCmd.Flags())
}

func ServerWorkload(cmd *cobra.Command, args []string) error {
	// Setup
	appCtx := cmd.Context()
	comps := new(application.ComponentManager)

	manager, err := components.WebServer(web.StaticFolderFlag.Value())
	if err != nil {
		return fmt.Errorf("error setting up: %w", err)
	}

	// Execute
	manager.Start()

	// Await termination signal
	<-appCtx.Done()

	// make a ctx specially for shutdown
	shutCtx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	// Shutdown
	berr := comps.BeforeShutdown(shutCtx)
	if berr != nil {
		log.Error("errors while performing pre shutdown calls", "error", berr)
	}

	manager.Stop(shutCtx)

	aerr := comps.AfterShutDown(shutCtx)
	if aerr != nil {
		log.Error("errors while performing post shutdown calls", "error", berr)
	}

	return errors.Join(berr, aerr)
}
