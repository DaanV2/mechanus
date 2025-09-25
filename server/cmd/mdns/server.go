package cmd_mdns

import (
	"github.com/DaanV2/mechanus/server/infrastructure/transport/mdns"
	"github.com/spf13/cobra"
)

// mdns/serverCmd represents the mdns/server command
var serverCmd = &cobra.Command{
	Use:  "server",
	RunE: serverCmdWorkload,
}

func init() {
	rootCmd.AddCommand(serverCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// config/pathCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// config/pathCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func serverCmdWorkload(cmd *cobra.Command, args []string) error {
	conf := mdns.GetServerConfig(8080)
	serv, err := mdns.NewServer(cmd.Context(), conf)
	if err != nil {
		return err
	}

	go serv.Listen()

	<-cmd.Context().Done()

	return nil
}
