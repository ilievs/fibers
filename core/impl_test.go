package core

import (
	"cmp"
	"math/rand/v2"
	"slices"
	"strconv"
	"sync"
	"testing"
)

type EmptyDevice struct {
	IdField string
}

func NewEmptyDevice() *EmptyDevice {
	return &EmptyDevice{strconv.Itoa(int(rand.Int64()))}
}

func NewEmptyDeviceWithId(deviceId string) *EmptyDevice {
	return &EmptyDevice{deviceId}
}

func (d* EmptyDevice) Id() string {
	return d.IdField
}

func (d* EmptyDevice) ListCommands() ([]Command, error) {
	return nil, nil
}

func (d* EmptyDevice) SendCommand(command *Command) error {
	return nil
}

func (d* EmptyDevice) GetState() *State {
	return nil
}

func (d* EmptyDevice) SubcribeToStateChanges() (chan *State, error) {
	return nil, nil
}

func (d* EmptyDevice) SubcribeToErrorChanges() (chan error, error) {
	return nil, nil
}


func TestAddDevice(t *testing.T) {

	var doneChan = make(chan int)
	var devManager = NewBasicDeviceManager()
	go func() {
		for range 10 {
			devManager.AddDevice(NewEmptyDevice())
		}
		doneChan <- 1
	}()

	go func() {
		for range 10 {
			devManager.AddDevice(NewEmptyDevice())
		}
		doneChan <- 1
	}()
	go func() {
		for range 10 {
			devManager.AddDevice(NewEmptyDevice())
		}
		doneChan <- 1
	}()

	<-doneChan
	<-doneChan
	<-doneChan

	actualDevCount := len(devManager.devicesById)
	if actualDevCount != 30 {
		t.Fatal("Expected 30 devices, but got", actualDevCount)
	}
}

func TestListDevices(t *testing.T) {
	
	var devManager = NewBasicDeviceManager()
	expected := make([]SimpleDevice, 0)
	devMutex := sync.Mutex{}
	addDeviceFunc := func() {
		d := EmptyDevice{strconv.Itoa(int(rand.Int64()))}
		devMutex.Lock()
		expected = append(expected, &d)
		devMutex.Unlock()
		devManager.AddDevice(&d)
	}
	
	var doneChan = make(chan int)
	go func() {
		for range 10 {
			addDeviceFunc()
		}
		doneChan <- 1
	}()

	go func() {
		for range 10 {
			addDeviceFunc()
		}
		doneChan <- 1
	}()
	go func() {
		for range 10 {
			addDeviceFunc()
		}
		doneChan <- 1
	}()

	<-doneChan
	<-doneChan
	<-doneChan

	slices.SortFunc(expected, func(a, b SimpleDevice) int {
		return cmp.Compare(a.Id(), b.Id())
	})
	
	actualDevices := devManager.ListDevices()
	slices.SortFunc(actualDevices, func(a, b SimpleDevice) int {
		return cmp.Compare(a.Id(), b.Id())
	})
	
	for i := 0; i < 30; i++ {
		if actualDevices[i] != expected[i] {
			t.Fatal("Expected devices:", expected, ", but got:", actualDevices)				
		}
	}
}

func TestRemoveDevice(t *testing.T) {
	var devManager = NewBasicDeviceManager()

	ids := make([]string, 0)
	for range 10 {
		ids = append(ids, strconv.Itoa(int(rand.Int64())))
	}

	for i := range 10 {
		devManager.AddDevice(NewEmptyDeviceWithId(ids[i]))
	}

	var doneChan = make(chan int)
	go func() {
		for range 10 {
			devManager.AddDevice(NewEmptyDevice())
		}
		doneChan <- 1
	}()

	go func() {
		for range 10 {
			devManager.AddDevice(NewEmptyDevice())
		}
		doneChan <- 1
	}()
	go func() {
		for i := range 10 {
			devManager.RemoveDevice(ids[i])
		}
		doneChan <- 1
	}()

	<-doneChan
	<-doneChan
	<-doneChan
	
	actualDevCount := len(devManager.devicesById)
	if actualDevCount != 20 {
		t.Fatal("Expected 20 devices, but got", actualDevCount)
	}
}
