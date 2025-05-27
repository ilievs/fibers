package broker

import (
	"log"
	"sync"
	"sync/atomic"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"

	. "github.com/ilievs/fibers/device"
)

type Subscription struct {
	topicFilter string
}

type MqttBrokerAndClient struct {
	server *mochi.Server
	subscriberIdCounter uint32
	subscriptionsById map[int]*Subscription
	devices map[string]SimpleDevice
	deviceLock sync.Mutex
}

func NewBroker() *MqttBrokerAndClient {
	// Create the new MQTT Server.
	server := mochi.New(&mochi.Options{
		InlineClient: true,
	})
	
	devices := make(map[string]SimpleDevice)

	return &MqttBrokerAndClient{
		server: server,
		subscriberIdCounter: 1,
		subscriptionsById: make(map[int]*Subscription),
		devices: devices,
	}
}


func (m *MqttBrokerAndClient) Start() error {

	options := auth.Options{
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
	}
	
	err := m.server.AddHook(new(AddNewDeviceHook), &options)
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
		callbackFn func(cl *mochi.Client, sub packets.Subscription, pk packets.Packet)) error {

	err := m.server.Subscribe(topicFilter, int(m.subscriberIdCounter), callbackFn)
	if err != nil {
		return err
	}
	
	m.subscriptionsById[int(m.subscriberIdCounter)] = &Subscription{topicFilter}
	
	atomic.AddUint32(&m.subscriberIdCounter, 1)
	
	return nil
}

func (m *MqttBrokerAndClient) Publish(topic string, payload []byte) error {
	err := m.server.Publish(topic, payload, true, 0)
	if err != nil {
		return err
	}

	return nil
}
