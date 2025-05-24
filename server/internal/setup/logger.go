package setup

import (
	"os"
	"time"

	"github.com/DaanV2/mechanus/server/mechanus/styles"
	"github.com/DaanV2/mechanus/server/pkg/config"
	"github.com/charmbracelet/log"
)

var (
	LoggerConfig     = config.New("logger").WithValidate(validateLogger)
	ReportCallerFlag = LoggerConfig.Bool("log.report-caller", false, "Whenever or not to output the file that outputs the log")
	LevelFlag        = LoggerConfig.String("log.level", "info", "The debug level, levels are: debug, info, warn, error, fatal")
	FormatFlag       = LoggerConfig.String("log.format", "text", "The format of the logging")
)

func validateLogger(c *config.Config) error {
	_, err := log.ParseLevel(c.GetString("log.level"))

	return err
}

func Logger() {
	logOptions := log.Options{
		TimeFormat:      time.DateTime,
		ReportCaller:    true,
		ReportTimestamp: true,
		Formatter:       log.TextFormatter,
	}

	// Initialize the default logger.
	logger := log.NewWithOptions(os.Stderr, logOptions)
	logger.SetStyles(styles.LogStyle())

	log.SetDefault(logger)
	updateLogger(
		logger,
		ReportCallerFlag.Value(),
		LevelFlag.Value(),
		FormatFlag.Value(),
	)
}

func UpdateLogger(reportCaller bool, level, format string) {
	defer log.Debug("setup the logger", "level", level, "format", format, "report-caller", reportCaller)
	updateLogger(log.Default(), reportCaller, level, format)
}

func updateLogger(logger *log.Logger, reportCaller bool, level, format string) {
	logger.SetReportCaller(reportCaller)

	// level
	if level != "" {
		l, err := log.ParseLevel(level)
		if err != nil {
			log.Fatal("invalid log level", "error", err)
		}

		logger.SetLevel(l)
	}

	// format
	switch format {
	default:
		log.Warn("unknown log format, falling back to text, expected text, json or logfmt", "format", format)

		fallthrough
	case "text", "":
		logger.SetFormatter(log.TextFormatter)
	case "json":
		logger.SetFormatter(log.JSONFormatter)
	case "logfmt":
		logger.SetFormatter(log.LogfmtFormatter)
	}
}