package inmemorytransport

import "github.com/telemetrytv/zephyr"

func (c *InMemoryTransport) AnnounceGateway(gatewayDescriptor *zephyr.GatewayDescriptor) error {
	for _, handler := range c.gatewayAnnounceHandlers {
		handler(gatewayDescriptor)
	}
	return nil
}

func (c *InMemoryTransport) BindGatewayAnnounce(handler func(gatewayDescriptor *zephyr.GatewayDescriptor)) error {
	c.gatewayAnnounceHandlers = append(c.gatewayAnnounceHandlers, handler)
	return nil
}

func (c *InMemoryTransport) UnbindGatewayAnnounce() error {
	c.gatewayAnnounceHandlers = nil
	return nil
}
