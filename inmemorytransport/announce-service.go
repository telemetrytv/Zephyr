package inmemorytransport

import "github.com/telemetrytv/zephyr"

func (c *InMemoryTransport) AnnounceService(serviceDescriptor *zephyr.ServiceDescriptor) error {
	for _, handler := range c.serviceAnnounceHandlers {
		handler(serviceDescriptor)
	}
	return nil
}

func (c *InMemoryTransport) BindServiceAnnounce(handler func(serviceDescriptor *zephyr.ServiceDescriptor)) error {
	c.serviceAnnounceHandlers = append(c.serviceAnnounceHandlers, handler)
	return nil
}

func (c *InMemoryTransport) UnbindServiceAnnounce() error {
	c.serviceAnnounceHandlers = nil
	return nil
}
