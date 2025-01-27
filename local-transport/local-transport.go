package localtransport

import (
	"net/http"

	"github.com/telemetrytv/zephyr"
)

type LocalTransport struct {
	gatewayAnnounceHandlers []func(gatewayDescriptor *zephyr.GatewayDescriptor)
	serviceAnnounceHandlers []func(serviceDescriptor *zephyr.ServiceDescriptor)
	dispatchHandlers        map[string]func(responseWriter http.ResponseWriter, request *http.Request)
}

var _ zephyr.Transport = &LocalTransport{}

func New() *LocalTransport {
	return &LocalTransport{
		dispatchHandlers: map[string]func(responseWriter http.ResponseWriter, request *http.Request){},
	}
}
