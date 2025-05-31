package mqtt

import (
	mochi "github.com/mochi-mqtt/server/v2"
)

type MochiClient struct {
	server *mochi.Server
}

func NewMochiClient(server *mochi.Server) *MochiClient {
	return &MochiClient{
		server,
	}
}

func (m *MochiClient) Publish(topic string, payload []byte) error {
	err := m.server.Publish(topic, payload, true, 0)
	if err != nil {
		return err
	}

	return nil
}

