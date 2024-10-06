package xio

import (
	"os"

	"github.com/charmbracelet/log"
)

func MakeDirAll(folder string) {
	// Check if the folder exists
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		// Folder does not exist, create it
		err := os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			log.Error("error during creation of the directory", "error", err, "folder", folder)
			return
		}
	}
}