package natstransport

import (
	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
	"github.com/vmihailenco/msgpack/v5"
)

func (c *NatsTransport) AnnounceGateway(gatewayDescriptor *zephyr.GatewayDescriptor) error {
	descriptorBuf, err := msgpack.Marshal(gatewayDescriptor)
	if err != nil {
		return err
	}
	return c.NatsConnection.Publish(namespace("gateway.announce"), descriptorBuf)
}

func (c *NatsTransport) BindGatewayAnnounce(handler func(gatewayDescriptor *zephyr.GatewayDescriptor)) error {
	subHandler := func(msg *nats.Msg) {
		gatewayDescriptorBuf := msg.Data
		gatewayDescriptor := &zephyr.GatewayDescriptor{}

		if err := msgpack.Unmarshal(gatewayDescriptorBuf, gatewayDescriptor); err != nil {
			panic(err)
		}

		handler(gatewayDescriptor)
	}

	gatewayAnnounceSub, err := c.NatsConnection.Subscribe(namespace("gateway.announce"), subHandler)
	if err != nil {
		return err
	}

	c.unbindGatewayAnnounce = func() error {
		return gatewayAnnounceSub.Unsubscribe()
	}

	return nil
}

func (c *NatsTransport) UnbindGatewayAnnounce() error {
	return c.unbindGatewayAnnounce()
}
