package logging

import (
	"io"
	// nolint:depguard // this is needed for libraries that implement slog
	"log/slog"

	"github.com/charmbracelet/log"
)

func Writer(opts ...log.StandardLogOptions) io.Writer {
	return ToWriter(log.Default(), opts...)
}

func ToWriter(logger *log.Logger, opts ...log.StandardLogOptions) io.Writer {
	return logger.StandardLog(opts...).Writer()
}

func Slog() *slog.Logger {
	return ToSlog(log.Default())
}

func ToSlog(logger *log.Logger) *slog.Logger {
	return slog.New(logger)
}
