package localtransport

import "net/http"

func (c *LocalTransport) Dispatch(serviceName string, responseWriter http.ResponseWriter, request *http.Request) error {
	if handler, ok := c.dispatchHandlers[serviceName]; ok {
		handler(responseWriter, request)
	}
	return nil
}

func (c *LocalTransport) BindDispatch(serviceName string, handler func(responseWriter http.ResponseWriter, request *http.Request)) error {
	c.dispatchHandlers[serviceName] = handler
	return nil
}

func (c *LocalTransport) UnbindDispatch(serviceName string) error {
	delete(c.dispatchHandlers, serviceName)
	return nil
}
