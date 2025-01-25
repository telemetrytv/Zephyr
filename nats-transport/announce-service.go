package natstransport

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
)

type ServiceAnnounce struct {
	GatewayNames      []string                  `json:"gatewayNames"`
	ServiceDescriptor *zephyr.ServiceDescriptor `json:"serviceDescriptor"`
}

func (c *NatsTransport) AnnounceService(serviceDescriptor *zephyr.ServiceDescriptor) error {
	serviceDescriptorBuf, err := json.Marshal(serviceDescriptor)
	if err != nil {
		return err
	}
	return c.NatsConnection.Publish(namespace("service.announce"), serviceDescriptorBuf)
}

func (c *NatsTransport) BindServiceAnnounce(handler func(serviceDescriptor *zephyr.ServiceDescriptor)) error {
	subHandler := func(msg *nats.Msg) {
		serviceDescriptorBuf := msg.Data
		serviceDescriptor := &zephyr.ServiceDescriptor{}

		if err := json.Unmarshal(serviceDescriptorBuf, serviceDescriptor); err != nil {
			panic(err)
		}

		handler(serviceDescriptor)
	}

	serviceAnnounceSub, err := c.NatsConnection.Subscribe(namespace("service.announce"), subHandler)
	if err != nil {
		return err
	}
	c.unbindServiceAnnounce = func() error {
		return serviceAnnounceSub.Unsubscribe()
	}

	return nil
}

func (c *NatsTransport) UnbindServiceAnnounce() error {
	return c.unbindServiceAnnounce()
}
