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
	env = EnvironmentNamer().Replace(env)

	logger := log.With("env", env, "flag", flag.Name)
	if err := viper.BindPFlag(flag.Name, flag); err != nil {
		logger.Fatal("error during binding of the flag to viper configuration")
	}

	if err := viper.BindEnv(env, flag.Name); err != nil {
		logger.Fatal("error binding the flag to the environment value")
	}

	flag.Usage += fmt.Sprintf(" (env: %s)", env)
}

func EnvironmentNamer() *strings.Replacer {
	v := make([]string, 0, len(illegal_env)*2)
	for _, i := range illegal_env {
		v = append(v, i, "_")
	}

	return strings.NewReplacer(v...)
}
