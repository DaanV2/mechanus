package cmd_maps

import (
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// config/importCmd represents the config/save command
var importCmd = &cobra.Command{
	Use:   "import [filepath]",
	Short: "Imports the given universal vtt map",
	Example: "mechanus maps import ./dungeon.dd2vtt",
	RunE:  importCmdWorkload,
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// config/pathCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// config/pathCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func importCmdWorkload(cmd *cobra.Command, args []string) error {
	v, _ := cmd.Flags().GetString("path")
	log.Debug("saving config file", "path", v)

	return nil
}
