package cmd

import (
	"syscall"

	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/internal/process"
	"github.com/DaanV2/mechanus/server/pkg/config"
	"github.com/DaanV2/mechanus/server/pkg/mdns"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Run: ServerWorkload,
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
	flags := serverCmd.Flags()
	config.MDNSFlags(flags)
	config.BindFlags(flags)
}

func ServerWorkload(cmd *cobra.Command, args []string) {
	discovery, err := mdns.NewDiscoverService(mdns.FromFlags(cmd.Flags()))
	if err != nil {
		log.Fatal("couldn't setup mdns discovery", "error", err)
	}

	defer logging.ReportIf("couldn't close mdns discovery", discovery.Close)

	process.AwaitSignal(syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
}
