package localtransport

import "github.com/telemetrytv/zephyr"

func (c *LocalTransport) AnnounceGateway(gatewayDescriptor *zephyr.GatewayDescriptor) error {
	for _, handler := range c.gatewayAnnounceHandlers {
		handler(gatewayDescriptor)
	}
	return nil
}

func (c *LocalTransport) BindGatewayAnnounce(handler func(gatewayDescriptor *zephyr.GatewayDescriptor)) error {
	c.gatewayAnnounceHandlers = append(c.gatewayAnnounceHandlers, handler)
	return nil
}

func (c *LocalTransport) UnbindGatewayAnnounce() error {
	c.gatewayAnnounceHandlers = nil
	return nil
}
