package main

import (
	"log"
	"sync/atomic"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
)

type Subscription struct {
	topicFilter string
}

type MqttBrokerAndClient struct {
	server *mqtt.Server
	subscriberIdCounter uint32
	subscriptionsById map[int]*Subscription
}

func newBroker() *MqttBrokerAndClient {
	// Create the new MQTT Server.
	server := mqtt.New(&mqtt.Options{
		InlineClient: true,
	})

	return &MqttBrokerAndClient{
		server: server,
		subscriberIdCounter: 0,
		subscriptionsById: map[int]*Subscription{},
	}
}

func (m *MqttBrokerAndClient) Start() error {

	err := m.server.AddHook(new(auth.Hook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{ // Auth disallows all by default
			  {Username: "iot1", Password: "password1", Allow: true},
			  {Remote: "127.0.0.1:*", Allow: true},
			  {Remote: "localhost:*", Allow: true},
			},
			ACL: auth.ACLRules{ // ACL allows all by default
			  {Remote: "127.0.0.1:*"}, // local superuser allow all
			  {
				// user melon can read and write to their own topic
				Username: "iot1", Filters: auth.Filters{
				  "iot1/#":   auth.ReadWrite,
				  "updates/#": auth.WriteOnly, // can write to updates, but can't read updates from others
				},
			  },
			  {
				// Otherwise, no clients have publishing permissions
				Filters: auth.Filters{
				  "#":         auth.ReadOnly,
				  "updates/#": auth.Deny,
				},
			  },
			},
		  },
	})
	if err != nil {
		log.Fatal(err)
	}
	
	// Create a TCP listener on a standard port.
	tcp := listeners.NewTCP(listeners.Config{ID: "t1", Address: ":1883"})
	err = m.server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := m.server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()
	
	return nil
}

func (m *MqttBrokerAndClient) Subscribe(topicFilter string,
		callbackFn func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet)) error {

	atomic.AddUint32(&m.subscriberIdCounter, 1)

	err := m.server.Subscribe(topicFilter, int(m.subscriberIdCounter), callbackFn)
	if err != nil {
		return err
	}

	m.subscriptionsById[int(m.subscriberIdCounter)] = &Subscription{topicFilter}
	return nil
}

func (m *MqttBrokerAndClient) Publish(topic string, payload []byte) error {
	err := m.server.Publish(topic, payload, true, 0)
	if err != nil {
		return err
	}

	return nil
}
