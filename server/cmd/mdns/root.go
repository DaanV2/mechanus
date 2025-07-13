package cmd_mdns

import (
	"github.com/spf13/cobra"
)

// mdns/rootCmd represents the mdns/root command
var rootCmd = &cobra.Command{
	Use:   "mdns",
	Short: "multi-cast dns commands",
}

func AddCommand(parent *cobra.Command) {
	parent.AddCommand(rootCmd)

}