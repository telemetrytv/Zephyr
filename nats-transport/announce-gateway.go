package natstransport

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
)

type GatewayAnnounce struct {
	GatewayName        string                      `json:"gatewayName"`
	ServiceDescriptors []*zephyr.ServiceDescriptor `json:"serviceDescriptors"`
}

func (c *NatsTransport) AnnounceGateway(gatewayDescriptor *zephyr.GatewayDescriptor) error {
	descriptorBuf, err := json.Marshal(gatewayDescriptor)
	if err != nil {
		return err
	}
	return c.NatsConnection.Publish(namespace("gateway.announce"), descriptorBuf)
}

func (c *NatsTransport) BindGatewayAnnounce(handler func(gatewayDescriptor *zephyr.GatewayDescriptor)) error {
	subHandler := func(msg *nats.Msg) {
		gatewayDescriptorBuf := msg.Data
		gatewayDescriptor := &zephyr.GatewayDescriptor{}

		if err := json.Unmarshal(gatewayDescriptorBuf, gatewayDescriptor); err != nil {
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
