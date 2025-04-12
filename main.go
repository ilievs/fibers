package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

func main() {
	// Create signals channel to run server until interrupted
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()
  
	mqttBrokerAndClient := newBroker()
	mqttBrokerAndClient.Start()

	// Demonstration of using an inline client to directly subscribe to a topic and receive a message when
	// that subscription is activated. The inline subscription method uses the same internal subscription logic
	// as used for external (normal) clients.
	go func() {
		// Inline subscriptions can also receive retained messages on subscription.
		_ = mqttBrokerAndClient.Publish("direct/retained", []byte("retained message"))
		_ = mqttBrokerAndClient.Publish("direct/alternate/retained", []byte("some other retained message"))

		// Subscribe to a filter and handle any received messages via a callback function.
		callbackFn := func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
			fmt.Println("inline client received message from subscription", "client", cl.ID, "subscriptionId", sub.Identifier, "topic", pk.TopicName, "payload", string(pk.Payload))
		}

		_ = mqttBrokerAndClient.Subscribe("direct/#", callbackFn)
		_ = mqttBrokerAndClient.Subscribe("direct/#", callbackFn)
	}()
	
	// Run server until interrupted
	<-done
  
	// Cleanup
}