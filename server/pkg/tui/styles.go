package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

func LogStyle() *log.Styles {
	styles := log.DefaultStyles()

	styles.Levels[log.DebugLevel] = styles.Levels[log.DebugLevel].SetString("DEBUG")
	styles.Levels[log.InfoLevel] = styles.Levels[log.InfoLevel].SetString("INFO")
	styles.Levels[log.WarnLevel] = styles.Levels[log.WarnLevel].SetString("WARN")
	styles.Levels[log.ErrorLevel] = styles.Levels[log.ErrorLevel].SetString("ERROR")
	styles.Levels[log.FatalLevel] = styles.Levels[log.FatalLevel].SetString("FATAL")

	styles.Keys["err"] = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	styles.Keys["error"] = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	styles.Values["error"] = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("204")).Faint(true)
	styles.Values["error"] = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("204")).Faint(true)

	return styles
}
