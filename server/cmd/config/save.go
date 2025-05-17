package cmd_config

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// config/saveCmd represents the config/save command
var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Saves the config, optionally to the specified path",
	RunE:  saveCmdWorkload,
}

func init() {
	rootCmd.AddCommand(saveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// config/pathCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// config/pathCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	saveCmd.Flags().StringP("path", "p", "", "The path to store the config file to")
}

func saveCmdWorkload(cmd *cobra.Command, args []string) error {
	v, _ := cmd.Flags().GetString("path")
	log.Debug("saving config file", "path", v)

	var err error
	if v == "" {
		err = viper.WriteConfig()
	} else {
		v, err = filepath.Abs(v)
		if err != nil {
			return fmt.Errorf("can't make the path an absolute path: %w", err)
		}

		err = viper.WriteConfigAs(v)
	}

	if err != nil {
		log.Fatal("couldn't save the config file", "error", err)
	}

	return nil
}
