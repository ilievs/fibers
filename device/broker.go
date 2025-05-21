package device

import (
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
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

	return &MqttBrokerAndClient{
		server: server,
		subscriberIdCounter: 1,
		subscriptionsById: make(map[int]*Subscription),
		devices: make(map[string]SimpleDevice),
		,
	}
}

func (m *MqttBrokerAndClient) AddDevice(d SimpleDevice) {

}

type AddNewDeviceHook struct {
	mochi.HookBase
	mqttClient *MqttBrokerAndClient
}

// OnSessionEstablished is called when a new client establishes a session (after OnConnect).
func (h *AddNewDeviceHook) OnSessionEstablished(cl *mochi.Client, pk packets.Packet) {
	dev, err := NewRelayDevice(h.mqttClient.server, cl.ID)
	if err != nil {
		log.Println("Failed to add new device with ID", cl.ID,
			"- Error:", err, "- Closing connection!")
		cl.Net.Conn.Close()
	}
	devices[cl.ID] = dev
}

// OnDisconnect is called when a client is disconnected for any reason.
func (h *AddNewDeviceHook) OnDisconnect(cl *mochi.Client, err error, expire bool) {
	// remove the device from the internal state of the server
	delete(devices, cl.ID)
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
