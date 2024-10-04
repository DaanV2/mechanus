package cmd

import (
	"github.com/DaanV2/mechanus/server/internal/setup"
	"github.com/DaanV2/mechanus/server/pkg/config"
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logconfing := config.GetSub[config.Logger]("log")

		setup.UpdateLogger(
			logconfing.ReportCaller,
			logconfing.Level,
			logconfing.Format,
		)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal("error during executing command", "error", err)
	}
}

func init() {
	pflags := rootCmd.PersistentFlags()
	config.LoggerFlags(pflags)
	config.BindFlags(pflags)
}
