package utilcmd

import (
	"github.com/spf13/cobra"
)

// utilCmd represents the util command
var utilCmd = &cobra.Command{
	Use:   "util",
	Short: "A bunch of util function to help manage",
}

func AddUtilCmd(parent *cobra.Command) {
	parent.AddCommand(utilCmd)
}