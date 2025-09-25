package process

import (
	"os"
	"os/signal"
)

// AwaitSignal will block until one of the given signal is received
func AwaitSignal(sig ...os.Signal) {
	sigs := make(chan os.Signal, 8)

	signal.Notify(sigs, sig...)
	defer signal.Stop(sigs)
	defer close(sigs)

	<-sigs
}
