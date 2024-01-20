package zephyr

import (
	"time"

	"github.com/RobertWHurst/navaros"
)

type ServiceDescriptor struct {
	Name             string                     `json:"name"`
	RouteDescriptors []*navaros.RouteDescriptor `json:"httpRouteDescriptors"`
	LastSeenAt       *time.Time                 `json:"-"`
	UnreachableAt    *time.Time                 `json:"-"`
	UnreachableCount int                        `json:"-"`
}
