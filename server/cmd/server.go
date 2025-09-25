package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"
	web "github.com/DaanV2/mechanus/server/infrastructure/transport/http"
	"github.com/DaanV2/mechanus/server/internal/checks"
	"github.com/DaanV2/mechanus/server/internal/components"
	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/networking/mdns"
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
	database.DatabaseConfig.AddToSet(serverCmd.Flags())
	checks.InitializeConfig.AddToSet(serverCmd.Flags())
	mdns.MDNSConfig.AddToSet(serverCmd.Flags())
}

func ServerWorkload(cmd *cobra.Command, args []string) error {
	// Setup
	server, err := components.BuildServer(cmd.Context())
	if err != nil {
		return fmt.Errorf("couldn't setup the server: %w", err)
	}
	// Start initial components
	err = server.Components.AfterInitialize(cmd.Context())
	if err != nil {
		log.Fatal("errors while performing initialization calls", "error", err)
	}

	checks.InitializeServer(cmd.Context(), server)
	manager := server.Manager

	// Execute
	manager.Start()

	// Await termination signal
	<-cmd.Context().Done()

	// make a ctx specially for shutdown
	shutCtx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()

	// Shutdown
	berr := server.Components.BeforeShutdown(shutCtx)
	if berr != nil {
		log.Error("errors while performing pre shutdown calls", "error", berr)
	}

	manager.Stop(shutCtx)

	aerr := server.Components.AfterShutDown(shutCtx)
	if aerr != nil {
		log.Error("errors while performing post shutdown calls", "error", berr)
	}

	return errors.Join(berr, aerr)
}
