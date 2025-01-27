package localtransport

import "github.com/telemetrytv/zephyr"

func (c *LocalTransport) AnnounceService(serviceDescriptor *zephyr.ServiceDescriptor) error {
	for _, handler := range c.serviceAnnounceHandlers {
		handler(serviceDescriptor)
	}
	return nil
}

func (c *LocalTransport) BindServiceAnnounce(handler func(serviceDescriptor *zephyr.ServiceDescriptor)) error {
	c.serviceAnnounceHandlers = append(c.serviceAnnounceHandlers, handler)
	return nil
}

func (c *LocalTransport) UnbindServiceAnnounce() error {
	c.serviceAnnounceHandlers = nil
	return nil
}
