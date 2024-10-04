package config

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var illegal_env = []string{
	"-", ".",
}

func BindFlags(flags *pflag.FlagSet) {
	flags.VisitAll(BindFlag)
}

func BindFlag(flag *pflag.Flag) {
	env := flag.Name
	env = strings.ToUpper(env)

	for _, old := range illegal_env {
		env = strings.ReplaceAll(env, old, "_")
	}

	logger := log.With("env", env, "flag", flag.Name)
	if err := viper.BindPFlag(flag.Name, flag); err != nil {
		logger.Fatal("error during binding of the flag to viper configuration")
	}
	if err := viper.BindEnv(env, flag.Name); err != nil {
		logger.Fatal("error binding the flag to the environment value")
	}

	flag.Usage += fmt.Sprintf(" (env: %s)", env)
}
