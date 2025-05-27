package main

import mqtt "github.com/mochi-mqtt/server/v2"

type Context struct {
	server mqtt.Server
}

func BuildContext() {
	mqttBrokerAndClient := mqtt.NewBroker()
}
