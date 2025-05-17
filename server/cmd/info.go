package cmd

import (
	"fmt"
	"log"

	"github.com/DaanV2/mechanus/server/mechanus/paths"
	"github.com/spf13/cobra"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "print info of mechanus",
	RunE:  PrintInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// infoCmd.PersistentFlags().String("foo", "", "A help for foo")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// infoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func PrintInfo(cmd *cobra.Command, args []string) error {
	log.Println("printing info")
	printInfoFn("app config dir", paths.GetAppConfigDir)
	printInfoFn("state dir", paths.GetStateDir)
	printInfoFn("user data dir", paths.GetUserDataDir)

	return nil
}

func printInfo(key, value any) {
	fmt.Println(key, "=", value)
}

func printInfoFn(key string, call func() (string, error)) {
	v, err := call()
	if err != nil {
		log.Fatal("error during reading of key/value", "key", key, "value", v, "error", err)
	}

	printInfo(key, v)
}
