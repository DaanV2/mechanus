package cmd

import (
	userscmd "github.com/DaanV2/mechanus/server/cmd/users"
	utilcmd "github.com/DaanV2/mechanus/server/cmd/util"
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
		setup.UpdateLogger(
			config.Logger.ReportCaller.Value(),
			config.Logger.Level.Value(),
			config.Logger.Format.Value(),
		)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer func ()  {
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
	config.Logger.AddToSet(pflags)

	utilcmd.AddCmds(rootCmd)
	userscmd.AddCmds(rootCmd)
}
