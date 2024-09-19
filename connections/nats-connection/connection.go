package natsconnection

import (
	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
)

type Connection struct {
	NatsConnection        *nats.Conn
	unbindDispatch        map[string][]func()
	unbindServiceAnnounce map[string][]func()
	unbindGatewayAnnounce map[string][]func()
}

func New(natsConnection *nats.Conn) zephyr.Connection {
	return &Connection{
		NatsConnection:        natsConnection,
		unbindDispatch:        map[string][]func(){},
		unbindServiceAnnounce: map[string][]func(){},
		unbindGatewayAnnounce: map[string][]func(){},
	}
}
