package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DaanV2/mechanus/server/application/checks"
	"github.com/DaanV2/mechanus/server/components"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/servers"
	"github.com/DaanV2/mechanus/server/infrastructure/tracing"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/cors"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/websocket"
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
	servers.ServerConfigSet.AddToSet(serverCmd.Flags())
	persistence.DatabaseConfigSet.AddToSet(serverCmd.Flags())
	checks.InitializeConfig.AddToSet(serverCmd.Flags())
	tracing.OtelConfigSet.AddToSet(serverCmd.Flags())
	cors.CorsConfig.AddToSet(serverCmd.Flags())
	websocket.WebsocketConfigSet.AddToSet(serverCmd.Flags())
}

// ServerWorkload starts and manages the server lifecycle.
func ServerWorkload(cmd *cobra.Command, args []string) error {
	// Setup
	cmpts, err := components.BuildServer(cmd.Context())
	if err != nil {
		return fmt.Errorf("couldn't setup the server: %w", err)
	}
	// Start initial components
	err = cmpts.Components.AfterInitialize(cmd.Context())
	if err != nil {
		log.Fatal("errors while performing initialization calls", "error", err)
	}

	checks.InitializeServer(cmd.Context(), cmpts)
	server := cmpts.Server

	// Execute
	go server.Listen()

	// Await termination signal
	<-cmd.Context().Done()

	// make a ctx specially for shutdown
	shutCtx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()

	// Shutdown
	berr := cmpts.Components.BeforeShutdown(shutCtx)
	if berr != nil {
		log.Error("errors while performing pre shutdown calls", "error", berr)
	}

	server.Shutdown(shutCtx)

	aerr := cmpts.Components.AfterShutDown(shutCtx)
	if aerr != nil {
		log.Error("errors while performing post shutdown calls", "error", aerr)
	}
	serr := cmpts.Components.ShutdownCleanup(shutCtx)
	if serr != nil {
		log.Error("errors while performing shutdown cleanup calls", "error", serr)
	}

	return errors.Join(berr, aerr, serr)
}
