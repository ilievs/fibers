package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ilievs/fibers/device"
)

func main() {

	mqttBrokerAndClient := device.NewBroker()
	mqttBrokerAndClient.Start()

	// Create signals channel to run server until interrupted
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()
  
	// Run server until interrupted
	<-done
  
	// Cleanup
}