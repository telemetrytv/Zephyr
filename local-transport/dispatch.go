package localtransport

import "net/http"

func (c *LocalTransport) Dispatch(serviceName string, responseWriter http.ResponseWriter, request *http.Request) error {
	transportLocalDispatchDebug.Tracef("Dispatching request to service %s: %s %s", 
		serviceName, request.Method, request.URL.Path)
	
	if handler, ok := c.dispatchHandlers[serviceName]; ok {
		transportLocalDispatchDebug.Tracef("Found handler for service %s, calling handler", serviceName)
		handler(responseWriter, request)
		transportLocalDispatchDebug.Tracef("Handler for service %s completed", serviceName)
	} else {
		transportLocalDispatchDebug.Tracef("No handler found for service %s", serviceName)
	}
	
	return nil
}

func (c *LocalTransport) BindDispatch(serviceName string, handler func(responseWriter http.ResponseWriter, request *http.Request)) error {
	transportLocalDispatchDebug.Tracef("Binding dispatch handler for service %s", serviceName)
	c.dispatchHandlers[serviceName] = handler
	return nil
}

func (c *LocalTransport) UnbindDispatch(serviceName string) error {
	transportLocalDispatchDebug.Tracef("Unbinding dispatch handler for service %s", serviceName)
	delete(c.dispatchHandlers, serviceName)
	return nil
}
