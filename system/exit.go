package system

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitForOsSignal() {
	// Create signals channel to run the app until interrupted
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	// Run app until interrupted
	<-done
}
