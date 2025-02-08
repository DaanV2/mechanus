package cmd

import (
	"context"
	"time"

	"github.com/DaanV2/mechanus/server/internal/web"
	"github.com/DaanV2/mechanus/server/pkg/application"
	"github.com/DaanV2/mechanus/server/pkg/servers"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Run:   ServerWorkload,
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
}

func ServerWorkload(cmd *cobra.Command, args []string) {
	// Setup
	appCtx := cmd.Context()
	comps := new(application.ComponentManager)

	router := web.WebRouter(comps, web.StaticFolderFlag.Value())
	webServer := servers.NewHttpServer(router, servers.HttpServerConfig{
		Port: web.PortFlag.Value(),
		Host: web.HostFlag.Value(),
	})

	// Execute
	go webServer.Listen()

	// Await termination signal
	<-appCtx.Done()

	// make a ctx specially for shutdown
	shutCtx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Shutdown
	if err := comps.BeforeShutdown(shutCtx); err != nil {
		log.Error("errors while performing pre shutdown calls", "error", err)
	}
	webServer.Shutdown()
	if err := comps.AfterShutDown(shutCtx); err != nil {
		log.Error("errors while performing post shutdown calls", "error", err)
	}
}
