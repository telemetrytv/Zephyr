package inmemorytransport

import (
	"net/http"

	"github.com/telemetrytv/zephyr"
)

type InMemoryTransport struct {
	gatewayAnnounceHandlers []func(gatewayDescriptor *zephyr.GatewayDescriptor)
	serviceAnnounceHandlers []func(serviceDescriptor *zephyr.ServiceDescriptor)
	dispatchHandlers        map[string]func(responseWriter http.ResponseWriter, request *http.Request)
}

var _ zephyr.Transport = &InMemoryTransport{}

func New() *InMemoryTransport {
	return &InMemoryTransport{
		dispatchHandlers: map[string]func(responseWriter http.ResponseWriter, request *http.Request){},
	}
}
