package mqtt

import (
	"log"
	"sync"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
)

type Subscription struct {
	topicFilter string
}

type MochiBroker struct {
	server              *mochi.Server
	subscriberIdCounter uint32
	subscriptionsById   map[int]*Subscription
	subscriberMutex sync.Mutex
}

func NewMochiBroker(server *mochi.Server) *MochiBroker {
	// Create the new MQTT Server.
	return &MochiBroker{
		server:              server,
		subscriberIdCounter: 1,
		subscriptionsById:   make(map[int]*Subscription),
	}
}

func (m *MochiBroker) Start(hooks []mochi.Hook, hookConfigs []any) error {

	options := auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{ // Auth disallows all by default
				{Username: "psu1", Password: "password1", Allow: true},
			},
			ACL: auth.ACLRules{ // ACL allows all by default
				{Remote: "127.0.0.1:*"}, // local superuser allow all
				{
					// user psu1 can read and write to their own topic
					Username: "psu1", Filters: auth.Filters{
						"devices/psu1/#":    auth.ReadWrite,
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

	err := m.server.AddHook(new(auth.Hook), &options)
	if err != nil {
		log.Fatal(err)
	}

	for i, hook := range hooks {
		err := m.server.AddHook(hook, hookConfigs[i])
		if err != nil {
			log.Fatal(err)
		}
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

func (m *MochiBroker) Subscribe(topicFilter string,
		callbackFn func(cl *mochi.Client, sub packets.Subscription, pk packets.Packet)) error {

	m.subscriberMutex.Lock()
	defer m.subscriberMutex.Unlock()
	err := m.server.Subscribe(topicFilter, int(m.subscriberIdCounter), callbackFn)
	if err != nil {
		return err
	}

	m.subscriptionsById[int(m.subscriberIdCounter)] = &Subscription{topicFilter}
	m.subscriberIdCounter += 1

	return nil
}
