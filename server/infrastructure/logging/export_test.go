package logging

import (
	"io"

	"github.com/charmbracelet/log"
)

// UseConsole points the console used by the bridge at w.
func UseConsole(w io.Writer) {
	console = newConsole(w)
	consoleOut = w
}

// ResetSink undoes [SetSink], the default logger itself has to be restored by the caller.
func ResetSink() {
	bridgeMu.Lock()
	defer bridgeMu.Unlock()

	bridged = false
}

// ConsoleFormatter changes the format of the console, as [UpdateLogger] would.
func ConsoleFormatter(f log.Formatter) {
	console.SetFormatter(f)
}
