package zephyr_test

import (
	"testing"

	"github.com/RobertWHurst/navaros"
	"github.com/telemetrytv/zephyr"
)

func Test_GatewayServiceIndexer_SetServiceDescriptor(t *testing.T) {
	r := &zephyr.GatewayServiceIndexer{}

	descriptor := &zephyr.ServiceDescriptor{
		Name: "service",
	}
	err := r.SetServiceDescriptor(descriptor)
	if err != nil {
		t.Errorf("SetServiceDescriptor should not return an error, got %v", err)
	}
	if len(r.ServiceDescriptors) != 1 {
		t.Errorf("ServiceDescriptors should have 1 item, got %v", len(r.ServiceDescriptors))
	}
	if r.ServiceDescriptors[0] != descriptor {
		t.Errorf("ServiceDescriptors[0] should be descriptor, got %v", r.ServiceDescriptors[0])
	}
}

func Test_GatewayServiceIndexer_UnsetService(t *testing.T) {
	r := &zephyr.GatewayServiceIndexer{}

	descriptor := &zephyr.ServiceDescriptor{
		Name: "service",
	}
	r.ServiceDescriptors = append(r.ServiceDescriptors, descriptor)

	err := r.UnsetService("service")
	if err != nil {
		t.Errorf("UnsetService should not return an error, got %v", err)
	}
	if len(r.ServiceDescriptors) != 0 {
		t.Errorf("ServiceDescriptors should have 0 items, got %v", len(r.ServiceDescriptors))
	}
}

func Test_GatewayServiceIndexer_ResolveService(t *testing.T) {
	r := &zephyr.GatewayServiceIndexer{}

	pattern, err := navaros.NewPattern("/path")
	if err != nil {
		t.Errorf("NewPattern should not return an error, got %v", err)
	}

	descriptor := &zephyr.ServiceDescriptor{
		Name: "service",
		RouteDescriptors: []*navaros.RouteDescriptor{
			{
				Method:  navaros.Get,
				Pattern: pattern,
			},
		},
	}
	r.ServiceDescriptors = append(r.ServiceDescriptors, descriptor)

	serviceName, ok := r.ResolveService(navaros.Get, "/path")
	if !ok {
		t.Errorf("ResolveService should return true")
	}
	if serviceName != "service" {
		t.Errorf("ResolveService should return service, got %v", serviceName)
	}
}
