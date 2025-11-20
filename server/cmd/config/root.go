package cmd_config

import (
	"github.com/spf13/cobra"
)

// config/rootCmd represents the config/root command
var rootCmd = &cobra.Command{
	Use:   "config",
	Short: "manage config settings",
}

// AddCommand adds the config command to the parent command.
func AddCommand(parent *cobra.Command) {
	parent.AddCommand(rootCmd)
}
