package config

import "github.com/spf13/pflag"

type LoggerConfig struct {
	ReportCaller Flag[bool]
	Level        Flag[string]
	Format       Flag[string]
}

var Logger = &LoggerConfig{
	ReportCaller: Bool("log.report-caller", false, "Whenever or not to output the file that outputs the log"),
	Level:        String("log.level", "info", "The debug level, levels are: debug, info, warn, error, fatal"),
	Format:       String("log.format", "text", "The text format of the logger"),
}

func (c *LoggerConfig) AddToSet(set *pflag.FlagSet) {
	c.ReportCaller.AddToSet(set)
	c.Level.AddToSet(set)
	c.Format.AddToSet(set)
}
