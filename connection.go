package zephyr

import (
	"net/http"
)

type Connection interface {
	AnnounceGateway(name string, serviceDescriptors []*ServiceDescriptor) error
	BindGatewayAnnounce(gatewayName string, handler func(serviceDescriptors []*ServiceDescriptor)) error
	UnbindGatewayAnnounce(gatewayName string)

	AnnounceService(gatewayName string, serviceDescriptor *ServiceDescriptor) error
	BindServiceAnnounce(gatewayName string, handler func(serviceDescriptor *ServiceDescriptor)) error
	UnbindServiceAnnounce(gatewayName string)

	Dispatch(serviceName string, res http.ResponseWriter, req *http.Request) error
	BindDispatch(serviceName string, handler func(res http.ResponseWriter, req *http.Request)) error
	UnbindDispatch(serviceName string)
}
