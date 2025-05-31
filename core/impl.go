package core

import (
	"sync"

	"github.com/ilievs/fibers/util"
)

type BasicDeviceManager struct {
	devicesById map[string]SimpleDevice
	devicesMutex sync.RWMutex
	subscribersMutex sync.RWMutex
	deviceAddedChannels []chan SimpleDevice
	stateChangeChannels []chan SimpleDevice
	errorChannels []chan error
}

func NewBasicDeviceManager() *BasicDeviceManager {
	return &BasicDeviceManager{
		devicesById: make(map[string]SimpleDevice),
		deviceAddedChannels: make([]chan SimpleDevice, 10),
		stateChangeChannels: make([]chan SimpleDevice, 10),
		errorChannels: make([]chan error, 10),
	}
}

func (m *BasicDeviceManager) AddDevice(d SimpleDevice) error {
	m.devicesMutex.Lock()
	defer m.devicesMutex.Unlock()
	m.devicesById[d.Id()] = d
	stateChan, err := d.SubcribeToStateChanges()
	if err != nil {
		return err
	}

	go func() {
		for range stateChan {
			for _, ch := range m.stateChangeChannels {
				go func() {
					ch <- d
				}()
			}
		}
	}()

	return nil
}

func (m *BasicDeviceManager) ListDevices() []SimpleDevice {
	m.devicesMutex.Lock()
	defer m.devicesMutex.Unlock()
	// TODO optimize returning a list of devices
	return util.Values(m.devicesById)
}

func (m *BasicDeviceManager) RemoveDevice(id string) {
	m.devicesMutex.Lock()
	defer m.devicesMutex.Unlock()
	delete(m.devicesById, id)
}

func (m *BasicDeviceManager) SubscribeToNewDeviceAdded() chan SimpleDevice {
	m.subscribersMutex.Lock()
	defer m.subscribersMutex.Unlock()
	newStateChan := make(chan SimpleDevice)
	m.deviceAddedChannels = append(m.deviceAddedChannels, newStateChan)
	return newStateChan
}

func (m *BasicDeviceManager) SubscribeToStateChanges() chan SimpleDevice {
	m.subscribersMutex.Lock()
	defer m.subscribersMutex.Unlock()
	newStateChangeChan := make(chan SimpleDevice)
	m.stateChangeChannels = append(m.stateChangeChannels, newStateChangeChan)
	return newStateChangeChan
}

func (m *BasicDeviceManager) SubscribeToErrors() chan error {
	m.subscribersMutex.Lock()
	defer m.subscribersMutex.Unlock()
	newErrorChan := make(chan error)
	m.errorChannels = append(m.errorChannels, newErrorChan)
	return newErrorChan
}

func (m *BasicDeviceManager) SendCommand(deviceId string, command *Command) error {
	return m.devicesById[deviceId].SendCommand(command)
}
