package setup

import (
	"os"
	"time"

	"github.com/DaanV2/mechanus/server/pkg/config"
	"github.com/charmbracelet/lipgloss"
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
		ReportCaller:    ReportCallerFlag.Value(),
		ReportTimestamp: true,
		Formatter:       log.TextFormatter,
	}

	// Initialize the default logger.
	logger := log.NewWithOptions(os.Stderr, logOptions)
	logger.SetStyles(CreateStyle())

	log.SetDefault(logger)
	UpdateLogger(
		ReportCallerFlag.Value(),
		LevelFlag.Value(),
		FormatFlag.Value(),
	)
}

func UpdateLogger(reportCaller bool, level, format string) {
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

  logger.Debug("setup the logger", "level", level, "format", format, "report-caller", reportCaller)
}

func CreateStyle() *log.Styles {
	styles := log.DefaultStyles()

	styles.Levels[log.DebugLevel] = styles.Levels[log.DebugLevel].SetString("DEBUG")
	styles.Levels[log.InfoLevel] = styles.Levels[log.InfoLevel].SetString("INFO")
	styles.Levels[log.WarnLevel] = styles.Levels[log.WarnLevel].SetString("WARN")
	styles.Levels[log.ErrorLevel] = styles.Levels[log.ErrorLevel].SetString("ERROR")
	styles.Levels[log.FatalLevel] = styles.Levels[log.FatalLevel].SetString("FATAL")

	styles.Keys["err"] = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	styles.Keys["error"] = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	styles.Values["error"] = lipgloss.NewStyle().Bold(true)
	styles.Values["error"] = lipgloss.NewStyle().Bold(true)

	return styles
}
