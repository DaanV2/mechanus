package xio

import (
	"os"

	"github.com/charmbracelet/log"
)

// MakeDirAll creates the given folder, with all its parents.
// Logs any potential error instead of return them
func MakeDirAll(folder string) {
	// Check if the folder exists
	if !DirExists(folder) {
		log.Info("creating folder: " + folder)

		// Folder does not exist, create it
		err := os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			log.Error("error during creation of the directory", "error", err, "folder", folder)
		}
	}
}

func DirExists(folder string) bool {
	_, err := os.Stat(folder)

	return !os.IsNotExist(err)
}
