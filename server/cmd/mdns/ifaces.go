package cmd_mdns

import (
	"net"

	"github.com/DaanV2/mechanus/server/pkg/terminal"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// mdns/ifacesCmd represents the mdns/server command
var ifacesCmd = &cobra.Command{
	Use:   "ifaces",
	Short: "list all aviable ifaces for mdns",
	RunE:  ifacesCmdWorkload,
}

func init() {
	rootCmd.AddCommand(ifacesCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// config/pathCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// config/pathCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func ifacesCmdWorkload(cmd *cobra.Command, args []string) error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	t := terminal.NewTable[net.Interface](ifaceRow)

	t.SetColumns(table.Column{Title: "name"}, table.Column{Title: "address"})
	t.AddItems(ifaces)

	t.AutoWidth()

	_, err = tea.NewProgram(t).Run()

	return err
}

func ifaceRow(item net.Interface) []string {
	return []string{
		item.Name,
		item.HardwareAddr.String(),
	}
}
