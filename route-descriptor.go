package zephyr

import "github.com/RobertWHurst/navaros"

// RouteDescriptor defines a route this service can handle. A route is a
// HTTP method, and a path matching pattern. It is used by the zephyr gateway
// to determine which service to dispatch a request to.
type RouteDescriptor = navaros.RouteDescriptor

// NewRouteDescriptor creates a new RouteDescriptor from a method and a path
// pattern. The pattern determines which URL path this route will match.
//
// To understand the pattern syntax, see the [navaros package](https://github.com/RobertWHurst/Navaros?tab=readme-ov-file#route-patterns).
func NewRouteDescriptor(method string, patternStr string) (*RouteDescriptor, error) {
	pattern, err := navaros.NewPattern(patternStr)
	if err != nil {
		return nil, err
	}
	return &RouteDescriptor{
		Method:  navaros.HTTPMethod(method),
		Pattern: pattern,
	}, nil
}
