package utilcmd

import (
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
	"github.com/spf13/cobra"
)

// util/mdnsCheckCmd represents the util/mdnsCheck command
var mdnsCheckCmd = &cobra.Command{
	Use:   "mdns-check",
	Short: "A brief description of your command",
	RunE: func(cmd *cobra.Command, args []string) error {
		wg := sync.WaitGroup{}
		wg.Add(1)

		// Make a channel for results and start listening
		entriesCh := make(chan *mdns.ServiceEntry, 4)
		go func() {
			defer wg.Done()

			for {
				entry, ok := <- entriesCh
				if !ok {
					return
				}

				fmt.Printf("Got new entry: %v\n", entry)
			}
		}()

		// Start the lookup
		params := mdns.DefaultParams("*")
		params.Entries = entriesCh
		params.DisableIPv6 = true
		params.Timeout = time.Second * 15

		err := mdns.Query(params)
		close(entriesCh)

		wg.Wait()
		return err
	},
}

func init() {
	utilCmd.AddCommand(mdnsCheckCmd)
}
