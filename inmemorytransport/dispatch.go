package inmemorytransport

import "net/http"

func (c *InMemoryTransport) Dispatch(serviceName string, responseWriter http.ResponseWriter, request *http.Request) error {
	if handler, ok := c.dispatchHandlers[serviceName]; ok {
		handler(responseWriter, request)
	}
	return nil
}

func (c *InMemoryTransport) BindDispatch(serviceName string, handler func(responseWriter http.ResponseWriter, request *http.Request)) error {
	c.dispatchHandlers[serviceName] = handler
	return nil
}

func (c *InMemoryTransport) UnbindDispatch(serviceName string) error {
	delete(c.dispatchHandlers, serviceName)
	return nil
}
