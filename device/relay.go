package device

import (
	"errors"
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
	mqttClient *mochi.Server
	availableCommands []Command
	stateTopic string
	commandTopic string
	state *State
	subscriberIdCounter atomic.Uint32
	callbackMutex sync.RWMutex
	stateChannels []chan State
	errorChannels []chan error
}

func NewRelayDevice(mqttClient *mochi.Server, deviceName string) (*RelayDevice, error) {
	
	stateTopic := fmt.Sprintf("device/%s/state", deviceName)
	commandTopic := fmt.Sprintf("device/%s/command", deviceName)

	dev := &RelayDevice{
		mqttClient,
		[]Command{
			{"power", []string{"on", "off"}},
		},
		stateTopic,
		commandTopic,
		&State{},
		atomic.Uint32{},
		sync.RWMutex{},
		[]chan State{},
		[]chan error{},
	}
	
	go func() {
		// Subscribe to the divice's state filter and fanout the state to the subscribers
		_ = mqttClient.Subscribe(stateTopic, int(dev.subscriberIdCounter.Add(1)), dev.fanoutState)
	}()
	
	return dev, nil
}

func (d *RelayDevice) fanoutState(cl *mochi.Client, sub packets.Subscription, pk packets.Packet) {
	log.Println("received message client", cl.ID,
		"subscriptionId", sub.Identifier,
		"topic", pk.TopicName,
		"voltage", strconv.Itoa(int(pk.Payload[0])),
		"current", strconv.Itoa(int(pk.Payload[1])))
	
	d.state = &State{map[string]string{
		"voltage": strconv.Itoa(int(pk.Payload[0])),
		"current": strconv.Itoa(int(pk.Payload[1])),
	}}

	d.callbackMutex.RLock()
	for _, c := range d.stateChannels {
		go func() {
			c <- *d.state
		}()
	}
	d.callbackMutex.RUnlock()
}

func Id() string {
	
}

func (d *RelayDevice) ListCommands() ([]Command, error) {
	return d.availableCommands, nil
}

func (d *RelayDevice) SendCommand(command Command) error {
	switch command.Name {
	case "power": {
		numArgs := len(command.Arguments)
		if numArgs != 1 {
			return fmt.Errorf("expected 1 argument. Instead got %d args", numArgs)
		}
		
		switch command.Arguments[0] {
		case "on":
			d.mqttClient.Publish("iot1/power", []byte{1}, true, 0)
			return nil
		case "off":
			d.mqttClient.Publish("iot1/power", []byte{0}, true, 0)
			return nil
		default:
			return fmt.Errorf("unknown command %s", command.Arguments[0])	
		}
	}
	default:
		return errors.New("unknown command")
	}
}

func (d *RelayDevice) SubcribeToStateChanges() (chan State, error) {
	d.callbackMutex.Lock()
	newStateChan := make(chan State)
	d.stateChannels = append(d.stateChannels, newStateChan)
	d.callbackMutex.Unlock()
	return newStateChan, nil
}

func (d *RelayDevice) SubcribeToErrorChanges() (chan error, error) {
	d.callbackMutex.Lock()
	newErrorChan := make(chan error)
	d.errorChannels = append(d.errorChannels, newErrorChan)
	d.callbackMutex.Unlock()
	return newErrorChan, nil
}
