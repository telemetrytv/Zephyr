package testconnection

import (
	"github.com/telemetrytv/zephyr"
)

func (c *Connection) AnnounceService(gatewayName string, serviceDescriptor *zephyr.ServiceDescriptor) error {
	return nil
}

func (c *Connection) BindGatewayAnnounce(gatewayName string, handler func(serviceDescriptors []*zephyr.ServiceDescriptor)) error {
	return nil
}

func (c *Connection) UnbindGatewayAnnounce(gatewayName string) {
}
