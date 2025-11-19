package cmd

import (
	"context"
	"os"
	"syscall"

	cmd_config "github.com/DaanV2/mechanus/server/cmd/config"
	cmd_maps "github.com/DaanV2/mechanus/server/cmd/maps"
	cmd_mdns "github.com/DaanV2/mechanus/server/cmd/mdns"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:                        "mechanus",
	Short:                      "🤖",
	SuggestionsMinimumDistance: 10,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logging.UpdateLogger(
			logging.ReportCallerFlag.Value(),
			logging.LevelFlag.Value(),
			logging.FormatFlag.Value(),
		)

		return logging.LoggerConfigSet.Validate()
	},
}

func init() {
	pflags := rootCmd.PersistentFlags()
	logging.LoggerConfigSet.AddToSet(pflags)

	cmd_config.AddCommand(rootCmd)
	cmd_mdns.AddCommand(rootCmd)
	cmd_maps.AddCommand(rootCmd)
}

// RootCommand returns the top level command of this package
func RootCommand() *cobra.Command {
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer func() {
		if e := recover(); e != nil {
			log.Fatal("uncaught error", "error", e)
		}
	}()

	err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithNotifySignal(syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT),
	)
	if err != nil {
		// nolint:gocritic // exitAfterDefer fine in this case, we already report the error
		log.Fatal("error during executing command", "error", err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
