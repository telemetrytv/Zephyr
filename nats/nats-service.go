package nats

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
)

func (c *Connection) AnnounceService(gatewayName string, serviceDescriptor *zephyr.ServiceDescriptor) error {
	descriptorBuf, err := json.Marshal(serviceDescriptor)
	if err != nil {
		return err
	}
	return c.Connection.Publish(namespace("service.announce", gatewayName), descriptorBuf)
}

func (c *Connection) BindGatewayAnnounce(gatewayName string, handler func(serviceDescriptors []*zephyr.ServiceDescriptor)) error {
	gatewayAnnounceSub, err := c.Connection.Subscribe(namespace("gateway.announce", gatewayName), func(msg *nats.Msg) {
		gatewayServiceDescriptorsBuf := msg.Data
		gatewayServiceDescriptors := []*zephyr.ServiceDescriptor{}

		if err := json.Unmarshal(gatewayServiceDescriptorsBuf, &gatewayServiceDescriptors); err != nil {
			panic(err)
		}

		handler(gatewayServiceDescriptors)
	})
	if err != nil {
		return err
	}

	c.onUnbindAnnounceGateway = func() {
		gatewayAnnounceSub.Unsubscribe()
	}

	return nil
}
