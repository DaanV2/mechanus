/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/DaanV2/mechanus/server/internal/terminal"
	"github.com/DaanV2/mechanus/server/pkg/config"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your application",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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

	rootCmd.SetHelpTemplate(terminal.HELP_TEMPLATE)
	rootCmd.SetUsageTemplate(terminal.USAGE_TEMPLATE)
}
