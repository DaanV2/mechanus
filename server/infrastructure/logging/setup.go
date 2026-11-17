package logging

import (
	"io"
	"os"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/config"
	"github.com/DaanV2/mechanus/server/pkg/tui"
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

var (
	LoggerConfigSet  = config.New("logger").WithValidate(validateLogger)
	ReportCallerFlag = LoggerConfigSet.Bool("log.report-caller", false, "Whenever or not to output the file that outputs the log")
	LevelFlag        = LoggerConfigSet.String("log.level", "info", "The debug level, levels are: debug, info, warn, error, fatal")
	FormatFlag       = LoggerConfigSet.String("log.format", "text", "The format of the logging")
)

func validateLogger(c *config.Config) error {
	_, err := log.ParseLevel(c.GetString("log.level"))

	return err
}

var (
	// console prints the lines of the default logger once a sink is set, see [SetSink]
	console    *log.Logger
	consoleOut io.Writer = os.Stderr
)

func SetupLogger() {
	logOptions := log.Options{
		TimeFormat:      time.DateTime,
		ReportCaller:    true,
		ReportTimestamp: true,
		Formatter:       log.TextFormatter,
	}

	// Initialize the default logger.
	logger := log.NewWithOptions(consoleOut, logOptions)
	logger.SetStyles(tui.LogStyle())

	console = newConsole(consoleOut)

	log.SetDefault(logger)
	updateLogger(
		logger,
		ReportCallerFlag.Value(),
		LevelFlag.Value(),
		FormatFlag.Value(),
	)

	viper.SetOptions(
		viper.WithLogger(Slog()),
	)
}

// newConsole creates the logger that prints the lines of a bridged default logger.
// Timestamp, level, caller and prefix are handed over by the bridge, so the console adds nothing itself.
func newConsole(w io.Writer) *log.Logger {
	c := log.NewWithOptions(w, log.Options{
		TimeFormat: time.DateTime,
		Formatter:  log.TextFormatter,
		Level:      log.DebugLevel,
	})
	c.SetStyles(tui.LogStyle())

	return c
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
	var formatter log.Formatter
	switch format {
	default:
		log.Warn("unknown log format, falling back to text, expected text, json or logfmt", "format", format)

		fallthrough
	case "text", "":
		formatter = log.TextFormatter
	case "json":
		formatter = log.JSONFormatter
	case "logfmt":
		formatter = log.LogfmtFormatter
	}

	if console != nil {
		console.SetFormatter(formatter)
	}
	// Once bridged the default logger has to stay JSON, the console takes over the format
	if !isBridged() {
		logger.SetFormatter(formatter)
	}
}
