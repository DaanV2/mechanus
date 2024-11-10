package utilcmd

import (
	"fmt"
	"net"

	"github.com/DaanV2/mechanus/server/pkg/terminal"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func init() {
	utilCmd.AddCommand(ifaceCmd)
}

// ifaceCmd represents the iface command
var ifaceCmd = &cobra.Command{
	Use:   "iface",
	Short: "List all the network interfaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		ifaces, err := net.Interfaces()
		if err != nil {
			return err
		}

		t := terminal.NewTable(func(item net.Interface) []string {
			return []string{
				item.Name,
				fmt.Sprintf("%v", item.Index),
				fmt.Sprintf("%v", item.MTU),
				item.HardwareAddr.String(),
				item.Flags.String(),
			}
		})
		t.SetColumns(
			table.Column{
				Title: "name",
				Width: 4,
			},
			table.Column{
				Title: "index",
				Width: 5,
			},
			table.Column{
				Title: "mtu",
				Width: 3,
			},
			table.Column{
				Title: "hardware address",
				Width: 16,
			},
			table.Column{
				Title: "flags",
				Width: 5,
			},
		)
		t.AddItems(ifaces)
		t.AutoWidth()

		_, err = tea.NewProgram(t).Run()
		return err
	},
}
