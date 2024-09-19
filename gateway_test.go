package zephyr_test

import (
	"testing"

	"github.com/telemetrytv/zephyr"
	testconnection "github.com/telemetrytv/zephyr/connections/test-connection"
)

func Test_NewGateway(t *testing.T) {
	connection := testconnection.New()

	gateway := zephyr.NewGateway("gateway", connection)

	if gateway.Name != "gateway" {
		t.Errorf("Gateway name should be 'gateway', got %v", gateway.Name)
	}
	if gateway.Connection != connection {
		t.Error("Gateway connection should be connection")
	}
}

func Test_Gateway_Start(t *testing.T) {
	connection := testconnection.New()

	gateway := zephyr.NewGateway("gateway", connection)

	if err := gateway.Start(); err != nil {
		t.Errorf("Start should not return an error, got %v", err)
	}

	testConnection := connection.(*testconnection.Connection)

	if !testConnection.CalledBindServiceAnnounce {
		t.Error("AnnounceGateway should be called")
	}
	if !testConnection.CalledAnnounceGateway {
		t.Error("AnnounceGateway should be called")
	}
}

func Test_Gateway_Stop(t *testing.T) {
	connection := testconnection.New()

	gateway := zephyr.NewGateway("gateway", connection)

	if err := gateway.Start(); err != nil {
		t.Errorf("Start should not return an error, got %v", err)
	}
	if err := gateway.Stop(); err != nil {
		t.Errorf("Stop should not return an error, got %v", err)
	}

	testConnection := connection.(*testconnection.Connection)

	if !testConnection.CalledUnbindServiceAnnounce {
		t.Error("UnbindServiceAnnounce should be called")
	}
}

func Test_Gateway_ServeHTTP(t *testing.T) {
	connection := testconnection.New()

	gateway := zephyr.NewGateway("gateway", connection)

	if err := gateway.Start(); err != nil {
		t.Errorf("Start should not return an error, got %v", err)
	}

	// FIXME: Implement this test
}
