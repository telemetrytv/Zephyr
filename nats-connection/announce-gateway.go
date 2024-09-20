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

func (c *Connection) BindServiceAnnounce(gatewayName string, handler func(serviceDescriptor *zephyr.ServiceDescriptor)) error {
	serviceAnnounceSub, err := c.NatsConnection.Subscribe(namespace("service.announce", gatewayName), func(msg *nats.Msg) {
		serviceDescriptorBuf := msg.Data
		serviceDescriptor := &zephyr.ServiceDescriptor{}

		if err := json.Unmarshal(serviceDescriptorBuf, &serviceDescriptor); err != nil {
			panic(err)
		}

		handler(serviceDescriptor)
	})
	if err != nil {
		return err
	}

	unbinders, ok := c.unbindServiceAnnounce[gatewayName]
	if !ok {
		unbinders = []func(){}
	}
	unbinders = append(unbinders, func() {
		serviceAnnounceSub.Unsubscribe()
	})
	c.unbindServiceAnnounce[gatewayName] = unbinders

	return nil
}

func (c *Connection) UnbindServiceAnnounce(gatewayName string) {
	unbinders, ok := c.unbindServiceAnnounce[gatewayName]
	if ok {
		for _, unbind := range unbinders {
			unbind()
		}
	}
}
