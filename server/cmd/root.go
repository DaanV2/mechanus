package cmd

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/DaanV2/mechanus/server/internal/setup"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mechanus",
	Short: "🤖",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		setup.UpdateLogger(
			setup.ReportCallerFlag.Value(),
			setup.LevelFlag.Value(),
			setup.FormatFlag.Value(),
		)

		return setup.LoggerConfig.Validate()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {	
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	defer cancel()
	rootCmd.SetContext(ctx)

	go func() {
		<- ctx.Done()
		log.Info("Shutdown received")
	}()

	defer func() {
		if e := recover(); e != nil {
			log.Fatal("uncaught error", "error", e)
		}
	}()

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal("error during executing command", "error", err)
	}
}

func init() {
	pflags := rootCmd.PersistentFlags()
	setup.LoggerConfig.AddToSet(pflags)
}