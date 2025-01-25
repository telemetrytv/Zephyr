package zephyr

type GatewayDescriptor struct {
	Name               string               `json:"name"`
	ServiceDescriptors []*ServiceDescriptor `json:"serviceDescriptors"`
}
