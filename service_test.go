package zephyr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/telemetrytv/zephyr"
	"github.com/telemetrytv/zephyr/inmemorytransport"
)

func TestService_Start(t *testing.T) {
	t.Run("Announces to gateway", func(t *testing.T) {
		conn := inmemorytransport.New()
		s := zephyr.NewService("testService", conn, nil)

		var serviceDescriptor *zephyr.ServiceDescriptor
		err := conn.BindServiceAnnounce(func(d *zephyr.ServiceDescriptor) {
			serviceDescriptor = d
		})
		assert.NoError(t, err)

		err = s.Start()
		assert.NoError(t, err)
		assert.Equal(t, "testService", serviceDescriptor.Name)
	})

	t.Run("Listens for gateway announce, and re-anounces itself if not known by the gateway", func(t *testing.T) {
		conn := inmemorytransport.New()
		s := zephyr.NewService("testService", conn, nil)

		err := s.Start()
		assert.NoError(t, err)

		var serviceDescriptor *zephyr.ServiceDescriptor
		err = conn.BindServiceAnnounce(func(d *zephyr.ServiceDescriptor) {
			serviceDescriptor = d
		})
		assert.NoError(t, err)

		gatewayDescriptor := &zephyr.GatewayDescriptor{
			Name: "testGateway",
		}
		err = conn.AnnounceGateway(gatewayDescriptor)
		assert.NoError(t, err)

		assert.Equal(t, "testService", serviceDescriptor.Name)
	})
}
