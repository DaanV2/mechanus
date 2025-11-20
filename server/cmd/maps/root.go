package cmd_maps

import (
	"github.com/spf13/cobra"
)

// maps/rootCmd represents the maps/root command
var rootCmd = &cobra.Command{
	Use:   "maps",
	Short: "manage maps",
}

// AddCommand adds the maps command to the parent command.
func AddCommand(parent *cobra.Command) {
	parent.AddCommand(rootCmd)
}
