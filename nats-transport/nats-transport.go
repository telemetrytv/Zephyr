package natstransport

import (
	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
)

type NatsTransport struct {
	NatsConnection        *nats.Conn
	unbindDispatch        map[string][]func() error
	unbindServiceAnnounce func() error
	unbindGatewayAnnounce func() error
}

var _ zephyr.Transport = &NatsTransport{}

func New(natsConnection *nats.Conn) *NatsTransport {
	return &NatsTransport{
		NatsConnection: natsConnection,
		unbindDispatch: map[string][]func() error{},
	}
}
