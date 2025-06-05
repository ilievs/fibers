package core

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type OnStateChangeCallback func(state State)
type OnErrorCallback func(err error)

type JsonCommDevice struct {
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

func NewRelayDevice(mqttClient *mochi.Server, deviceId string, deviceCommand []Command) (*JsonCommDevice, error) {
	
	stateTopic := fmt.Sprintf("devices/%s/state", deviceId)
	commandTopic := fmt.Sprintf("devices/%s/command", deviceId)

	dev := &JsonCommDevice{
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

func (d *JsonCommDevice) fanoutState(cl *mochi.Client, sub packets.Subscription, pk packets.Packet) {
	
	stateProperties := make(map[string]string)
	json.Unmarshal(pk.Payload, &stateProperties)

	log.Println("received message from client", cl.ID,
		"subscriptionId", sub.Identifier,
		"topic", pk.TopicName,
		"state", stateProperties)
	
	d.subscribersMutex.RLock()
	d.state = &State{stateProperties}
	for _, c := range d.stateChannels {
		go func() {
			c <- d.state
		}()
	}
	d.subscribersMutex.RUnlock()
}

func (d *JsonCommDevice) Id() string {
	return d.id
}

func (d *JsonCommDevice) ListCommands() ([]Command, error) {
	return d.availableCommands, nil
}

func (d *JsonCommDevice) SendCommand(command *Command) error {
	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}

	err = d.mqttClient.Publish(d.commandTopic, payload, true, 0)
	return err
}

func (d *JsonCommDevice) GetState() *State {
	return d.state
}

func (d *JsonCommDevice) SubcribeToStateChanges() (chan *State, error) {
	d.subscribersMutex.Lock()
	newStateChan := make(chan *State)
	d.stateChannels = append(d.stateChannels, newStateChan)
	d.subscribersMutex.Unlock()
	return newStateChan, nil
}

func (d *JsonCommDevice) SubcribeToErrorChanges() (chan error, error) {
	d.subscribersMutex.Lock()
	newErrorChan := make(chan error)
	d.errorChannels = append(d.errorChannels, newErrorChan)
	d.subscribersMutex.Unlock()
	return newErrorChan, nil
}
