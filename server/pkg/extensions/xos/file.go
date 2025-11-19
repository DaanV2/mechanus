package xos

import (
	"os"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xio"
	"github.com/charmbracelet/log"
)

const (
	DEFAULT_FILE_PERMISSIONS = xio.DEFAULT_FILE_PERMISSIONS
)

// WriteFile writes data to the named file with default file permissions.
func WriteFile(name string, data []byte) error {
	return os.WriteFile(name, data, xio.DEFAULT_FILE_PERMISSIONS)
}

// CloseOrReport will take the given file and stop it, if an error occurs it logger to the given logger.
// If the given logger is nil, the default [log.Default] is used
func CloseOrReport(f *os.File, logger *log.Logger) {
	err := f.Close()
	if err == nil {
		return
	}

	if logger == nil {
		logger = log.Default()
	}

	logger.With("error", err).Errorf("error closing the file: %s", f.Name())
}
