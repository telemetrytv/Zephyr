package testconnection

import (
	"github.com/telemetrytv/zephyr"
)

type Connection struct {
	CalledAnnounceGateway       bool
	CalledBindServiceAnnounce   bool
	CalledUnbindServiceAnnounce bool

	unbindDispatch        map[string][]func()
	unbindServiceAnnounce map[string][]func()
	unbindGatewayAnnounce map[string][]func()
}

func New() zephyr.Connection {
	return &Connection{
		unbindDispatch:        map[string][]func(){},
		unbindServiceAnnounce: map[string][]func(){},
		unbindGatewayAnnounce: map[string][]func(){},
	}
}
