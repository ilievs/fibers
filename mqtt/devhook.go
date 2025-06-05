package mqtt

import (
	"bytes"
	"log"

	mochi "github.com/mochi-mqtt/server/v2"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"

	"github.com/ilievs/fibers/core"
)

type HookOptions struct {
	MqttClient *MochiClient
	DeviceManager core.DeviceManager
}

type AddNewDeviceHook struct {
	mochi.HookBase
	mqttClient *MochiClient
	devMan     core.DeviceManager
}

// ID returns the ID of the hook.
func (h *AddNewDeviceHook) ID() string {
	return "AddNewDeviceHook"
}

// Provides indicates which methods a hook provides. The default is none - this method
// should be overridden by the embedding hook.
func (h *AddNewDeviceHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mochi.OnSessionEstablished,
		mqtt.OnDisconnect,
	}, []byte{b})
}

// Init performs any pre-start initializations for the hook, such as connecting to databases
// or opening files.
func (h *AddNewDeviceHook) Init(config any) error {
	if _, ok := config.(*HookOptions); !ok && config != nil {
		return mochi.ErrInvalidConfigType
	}

	if config == nil {
		config = new(HookOptions)
	}

	opt := config.(*HookOptions)
	h.mqttClient = opt.MqttClient
	h.devMan = opt.DeviceManager

	return nil
}

// OnSessionEstablished is called when a new client establishes a session (after OnConnect).
func (h *AddNewDeviceHook) OnSessionEstablished(cl *mochi.Client, pk packets.Packet) {
	deviceCommands := []core.Command{
		{Name: "power", Arguments: []string{"on", "off"}},
	}

	dev, err := core.NewRelayDevice(h.mqttClient.server, cl.ID, deviceCommands)
	if err != nil {
		log.Println("Failed to add new device with ID", cl.ID,
			"- Error:", err, "- Closing connection!")
		cl.Net.Conn.Close()
		// TODO publish the error event somewhere
		return
	}
	h.devMan.AddDevice(dev)

	log.Println("New device added", cl.ID)
}

// OnDisconnect is called when a client is disconnected for any reason.
func (h *AddNewDeviceHook) OnDisconnect(cl *mochi.Client, err error, expire bool) {
	// remove the device from the internal state of the server
	h.devMan.RemoveDevice(cl.ID)
	log.Println("Device removed", cl.ID)
}
