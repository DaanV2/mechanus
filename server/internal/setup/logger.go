package setup

import (
	"os"
	"time"

	"github.com/DaanV2/mechanus/server/pkg/config"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

func Logger() {
	logOptions := log.Options{
		TimeFormat:      time.DateTime,
		ReportCaller:    false,
		ReportTimestamp: true,
		Formatter:       log.TextFormatter,
	}

	// Initialize the default logger.
	logger := log.NewWithOptions(os.Stderr, logOptions)
	logger.SetStyles(CreateStyle())

	log.SetDefault(logger)
	UpdateLogger(
		config.Logger.ReportCaller.Value(),
		config.Logger.Level.Value(),
		config.Logger.Format.Value(),
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