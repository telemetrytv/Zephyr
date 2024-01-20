package nats

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
	return c.Connection.Publish(namespace("gateway.announce", gatewayName), descriptorsBuf)
}

func (c *Connection) BindServiceAnnounce(gatewayName string, handler func(serviceDescriptor *zephyr.ServiceDescriptor)) error {
	serviceAnnounceSub, err := c.Connection.Subscribe(namespace("service.announce", gatewayName), func(msg *nats.Msg) {
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

	c.onUnbindAnnounceService = func() {
		serviceAnnounceSub.Unsubscribe()
	}

	return nil
}

func (c *Connection) UnbindServiceAnnounce() {
	if c.onUnbindAnnounceService != nil {
		c.onUnbindAnnounceService()
	}
}
