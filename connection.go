package zephyr

import "github.com/RobertWHurst/navaros"

type Connection interface {
	AnnounceGateway(name string, serviceDescriptors []*ServiceDescriptor) error     // to service
	AnnounceService(gatewayName string, serviceDescriptor *ServiceDescriptor) error // to gateway
	DispatchToService(serviceName string, ctx *navaros.Context) error               // to and from service

	BindGatewayAnnounce(gatewayName string, handler func(serviceDescriptors []*ServiceDescriptor)) error // from gateway
	BindServiceAnnounce(gatewayName string, handler func(serviceDescriptor *ServiceDescriptor)) error    // from service
	BindDispatchFromGatewayOrService(serviceName string, handler func(ctx *navaros.Context)) error       // from gateway

	UnbindGatewayAnnounce()
	UnbindServiceAnnounce()
	UnBindDispatchFromGatewayOrService()
}
