package userscmd

import (
	"github.com/spf13/cobra"
)

// usersCmd represents the util command
var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "A bunch of function to help manage users",
}

func AddCmds(parent *cobra.Command) {
	parent.AddCommand(usersCmd)
}