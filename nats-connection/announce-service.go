package natsconnection

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
)

func (c *Connection) AnnounceService(gatewayName string, serviceDescriptor *zephyr.ServiceDescriptor) error {
	println("announcing service")

	descriptorBuf, err := json.Marshal(serviceDescriptor)
	if err != nil {
		return err
	}
	return c.NatsConnection.Publish(namespace("service.announce", gatewayName), descriptorBuf)
}

func (c *Connection) BindGatewayAnnounce(gatewayName string, handler func(serviceDescriptors []*zephyr.ServiceDescriptor)) error {
	gatewayAnnounceSub, err := c.NatsConnection.Subscribe(namespace("gateway.announce", gatewayName), func(msg *nats.Msg) {
		println("got gateway announce")

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

	unbinders, ok := c.unbindGatewayAnnounce[gatewayName]
	if !ok {
		unbinders = []func(){}
	}
	unbinders = append(unbinders, func() {
		gatewayAnnounceSub.Unsubscribe()
	})
	c.unbindGatewayAnnounce[gatewayName] = unbinders

	return nil
}

func (c *Connection) UnbindGatewayAnnounce(gatewayName string) {
	unbinders, ok := c.unbindGatewayAnnounce[gatewayName]
	if ok {
		for _, unbind := range unbinders {
			unbind()
		}
	}
}
