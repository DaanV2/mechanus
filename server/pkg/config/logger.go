package config

import "github.com/spf13/pflag"

type Logger struct {
	ReportCaller bool   `mapstructure:"log.report-caller"`
	Level        string `mapstructure:"log.level"`
	Format       string `mapstructure:"log.format"`
}

func LoggerFlags(flags *pflag.FlagSet) {
	flags.Bool("log.report-caller", false, "Whenever or not to output the file that outputs the log")
	flags.String("log.level", "info", "The debug level, levels are: debug, info, warn, error, fatal")
	flags.String("log.format", "text", "The text format of the logger")
}
