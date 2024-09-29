package setup

import (
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

func Logger() {
	logOptions := log.Options{
		TimeFormat:      time.DateTime,
		ReportCaller:    false,
		ReportTimestamp: false,
		Formatter:       log.TextFormatter,
	}

	// Initialize the default logger.
	logger := log.NewWithOptions(os.Stderr, logOptions)
	logger.SetStyles(CreateStyle())

	log.SetDefault(logger)
}

func UpdateLogger(reportCaller bool, level, format string) {
	logger := log.Default()
	updateLogger(logger, reportCaller, level, format)
}

func updateLogger(logger *log.Logger, reportCaller bool, level, format string) {
	logger.SetReportCaller(reportCaller)

	// level
	l, err := log.ParseLevel(level)
	if err != nil {
		log.Fatal("invalid log level", "error", err)
	}
	logger.SetLevel(l)

	// format
	switch format {
	default:
		log.Warn("unknown log format, falling back to text, expected text, json or logfmt", "format", format)
		fallthrough
	case "text":
		logger.SetFormatter(log.TextFormatter)
	case "json":
		logger.SetFormatter(log.JSONFormatter)
	case "logfmt":
		logger.SetFormatter(log.LogfmtFormatter)
	}
}

func CreateStyle() *log.Styles {
	styles := log.DefaultStyles()

	styles.Levels[log.DebugLevel] = styles.Levels[log.DebugLevel].SetString("🔎")
	styles.Levels[log.InfoLevel] = styles.Levels[log.InfoLevel].SetString("")
	styles.Levels[log.WarnLevel] = styles.Levels[log.WarnLevel].SetString("⚠️")
	styles.Levels[log.ErrorLevel] = styles.Levels[log.ErrorLevel].SetString("💥")
	styles.Levels[log.FatalLevel] = styles.Levels[log.FatalLevel].SetString("☠️")

	styles.Keys["err"] = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	styles.Keys["error"] = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	styles.Values["error"] = lipgloss.NewStyle().Bold(true)
	styles.Values["error"] = lipgloss.NewStyle().Bold(true)

	return styles
}
