package core

type DeviceManager interface {

	AddDevice(d SimpleDevice) error

	ListDevices() []SimpleDevice

	RemoveDevice(id string)

	SubscribeToNewDeviceAdded() chan SimpleDevice

	SubscribeToStateChanges() chan SimpleDevice
	
	SubscribeToErrors() chan error

	SendCommand(deviceId string, command *Command) error
}
