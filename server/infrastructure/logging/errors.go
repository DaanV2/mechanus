package logging

import "github.com/charmbracelet/log"

func ReportIf(msg string, call func() error) {
	err := call()
	if err != nil {
		log.Error(msg, "error", err)
	}
}
