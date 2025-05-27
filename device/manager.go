package device

type DeviceManager interface {

	AddDevice(d SimpleDevice)

	ListDevices() []SimpleDevice

	RemoveDevice(id string)

	SubscribeToNewDeviceAdded() chan SimpleDevice

	SubscribeToDeviceStateChanges() chan SimpleDevice
}
