package cmd

import (
	"fmt"
	"sort"

	"github.com/DaanV2/mechanus/server/pkg/paths"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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

// PrintInfo prints information about application directories and paths.
func PrintInfo(cmd *cobra.Command, args []string) error {
	values := []string{
		printInfoFn("app config dir", paths.GetAppConfigDir),
		printInfoFn("state dir", paths.GetStateDir),
		printInfoFn("user data dir", paths.GetUserDataDir),
	}
	// TODO use bubbles
	fmt.Println("printing info")

	fmt.Println("\n==== Info ====")
	sort.Strings(values)
	for _, v := range values {
		fmt.Println(v)
	}

	data, _ := yaml.Marshal(viper.AllSettings())
	fmt.Println(string(data))

	return nil
}

func printInfoFn(key string, call func() (string, error)) string {
	v, err := call()
	if err != nil {
		log.Fatal("error during reading of key/value", "key", key, "value", v, "error", err)
	}

	return fmt.Sprintf("%s=%s", key, v)
}
