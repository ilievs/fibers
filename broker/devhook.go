package broker

import (
	"log"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"

	. "github.com/ilievs/fibers/device"
)

type AddNewDeviceHook struct {
	mochi.HookBase
	mqttClient *MqttBrokerAndClient
	devMan DeviceManager
}

// OnSessionEstablished is called when a new client establishes a session (after OnConnect).
func (h *AddNewDeviceHook) OnSessionEstablished(cl *mochi.Client, pk packets.Packet) {
	dev, err := NewRelayDevice(h.mqttClient.server, cl.ID)
	if err != nil {
		log.Println("Failed to add new device with ID", cl.ID,
			"- Error:", err, "- Closing connection!")
		cl.Net.Conn.Close()
	}
	h.devMan.AddDevice(dev)
}

// OnDisconnect is called when a client is disconnected for any reason.
func (h *AddNewDeviceHook) OnDisconnect(cl *mochi.Client, err error, expire bool) {
	// remove the device from the internal state of the server
	h.devMan.RemoveDevice(cl.ID)
}
