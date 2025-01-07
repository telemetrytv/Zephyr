package natsconnection

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
)

func (c *Connection) AnnounceGateway(gatewayName string, serviceDescriptors []*zephyr.ServiceDescriptor) error {
	descriptorsBuf, err := json.Marshal(serviceDescriptors)
	if err != nil {
		return err
	}
	return c.NatsConnection.Publish(namespace("gateway.announce", gatewayName), descriptorsBuf)
}

func (c *Connection) BindGatewayAnnounce(gatewayName string, handler func(serviceDescriptors []*zephyr.ServiceDescriptor)) error {
	subHandler := func(msg *nats.Msg) {
		gatewayServiceDescriptorsBuf := msg.Data
		gatewayServiceDescriptors := []*zephyr.ServiceDescriptor{}

		if err := json.Unmarshal(gatewayServiceDescriptorsBuf, &gatewayServiceDescriptors); err != nil {
			panic(err)
		}

		handler(gatewayServiceDescriptors)
	}

	gatewayAnnounceSub, err := c.NatsConnection.Subscribe(namespace("gateway.announce"), subHandler)
	if err != nil {
		return err
	}
	namespacedGatewayAnnounceSub, err := c.NatsConnection.Subscribe(namespace("gateway.announce", gatewayName), subHandler)
	if err != nil {
		return err
	}

	unbinders, ok := c.unbindGatewayAnnounce[gatewayName]
	if !ok {
		unbinders = []func(){}
	}
	unbinders = append(unbinders, func() {
		if err := gatewayAnnounceSub.Unsubscribe(); err != nil {
			panic(err)
		}
		if err := namespacedGatewayAnnounceSub.Unsubscribe(); err != nil {
			panic(err)
		}
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
