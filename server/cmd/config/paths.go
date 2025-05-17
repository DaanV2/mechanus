package cmd_config

import (
	"fmt"

	"github.com/DaanV2/mechanus/server/pkg/config"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// config/pathCmd represents the config/paths command
var pathCmd = &cobra.Command{
	Use:   "paths",
	Short: "Output all the config folders",
	RunE:  pathsCmdWorkload,
}

func init() {
	rootCmd.AddCommand(pathCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// config/pathCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// config/pathCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func pathsCmdWorkload(cmd *cobra.Command, args []string) error {
	log.Info("config paths:")
	for _, p := range config.ConfigPaths() {
		fmt.Println("- ", p)
	}

	return nil
}
