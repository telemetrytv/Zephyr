package testconnection

import (
	"github.com/telemetrytv/zephyr"
)

func (c *Connection) AnnounceGateway(gatewayName string, serviceDescriptors []*zephyr.ServiceDescriptor) error {
	c.CalledAnnounceGateway = true
	return nil
}

func (c *Connection) BindServiceAnnounce(gatewayName string, handler func(serviceDescriptor *zephyr.ServiceDescriptor)) error {
	c.CalledBindServiceAnnounce = true
	return nil
}

func (c *Connection) UnbindServiceAnnounce(gatewayName string) {
	c.CalledUnbindServiceAnnounce = true
}
