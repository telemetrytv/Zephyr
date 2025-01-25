package zephyr

type GatewayDescriptor struct {
	Name               string               `msgpack:"name"`
	ServiceDescriptors []*ServiceDescriptor `msgpack:"serviceDescriptors"`
}
