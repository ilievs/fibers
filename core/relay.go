package core

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type OnStateChangeCallback func(state State)
type OnErrorCallback func(err error)

type RelayDevice struct {
	id string
	mqttClient *mochi.Server
	availableCommands []Command
	stateTopic string
	commandTopic string
	state *State
	subscriberIdCounter atomic.Uint32
	subscribersMutex sync.RWMutex
	stateChannels []chan *State
	errorChannels []chan error
}

func NewRelayDevice(mqttClient *mochi.Server, deviceId string, deviceCommand []Command) (*RelayDevice, error) {
	
	stateTopic := fmt.Sprintf("devices/%s/state", deviceId)
	commandTopic := fmt.Sprintf("devices/%s/command", deviceId)

	dev := &RelayDevice{
		id: deviceId,
		mqttClient: mqttClient,
		availableCommands: deviceCommand,
		stateTopic: stateTopic,
		commandTopic: commandTopic,
		state: &State{},
	}
	
	go func() {
		// Subscribe to the divice's state filter and fanout the state to the subscribers
		_ = mqttClient.Subscribe(stateTopic, int(dev.subscriberIdCounter.Add(1)), dev.fanoutState)
	}()
	
	return dev, nil
}

func (d *RelayDevice) fanoutState(cl *mochi.Client, sub packets.Subscription, pk packets.Packet) {
	voltage := strconv.Itoa(int(pk.Payload[0]))
	current := strconv.Itoa(int(pk.Payload[1]))

	log.Println("received message client", cl.ID,
		"subscriptionId", sub.Identifier,
		"topic", pk.TopicName,
		"voltage", voltage,
		"current", current)
	
	d.state = &State{
		map[string]string{
				"voltage": voltage,
				"current": current,
			},
		}

	d.subscribersMutex.RLock()
	for _, c := range d.stateChannels {
		go func() {
			c <- d.state
		}()
	}
	d.subscribersMutex.RUnlock()
}

func (d *RelayDevice) Id() string {
	return d.id
}

func (d *RelayDevice) ListCommands() ([]Command, error) {
	return d.availableCommands, nil
}

func (d *RelayDevice) SendCommand(command *Command) error {
	var payload []byte
	switch command.Name {
	case "power": {
		numArgs := len(command.Arguments)
		if numArgs != 1 {
			return fmt.Errorf("expected 1 argument. Instead got %d args", numArgs)
		}
		
		switch command.Arguments[0] {
		case "on":
			payload = []byte{1}
		case "off":
			payload = []byte{0}
		default:
			return fmt.Errorf("unknown command argument %s", command.Arguments[0])	
		}
	}
	default:
		return fmt.Errorf("unknown command %s", command.Name)
	}
	
	err := d.mqttClient.Publish(d.commandTopic, payload, true, 0)
	return err
}

func (d *RelayDevice) GetState() *State {
	return d.state
}

func (d *RelayDevice) SubcribeToStateChanges() (chan *State, error) {
	d.subscribersMutex.Lock()
	newStateChan := make(chan *State)
	d.stateChannels = append(d.stateChannels, newStateChan)
	d.subscribersMutex.Unlock()
	return newStateChan, nil
}

func (d *RelayDevice) SubcribeToErrorChanges() (chan error, error) {
	d.subscribersMutex.Lock()
	newErrorChan := make(chan error)
	d.errorChannels = append(d.errorChannels, newErrorChan)
	d.subscribersMutex.Unlock()
	return newErrorChan, nil
}
