package zephyr

import (
	"fmt"
	"net/http"

	"github.com/RobertWHurst/navaros"
)

type Service struct {
	GatewayNames     []string
	Name             string
	Connection       Connection
	RouteDescriptors []*navaros.RouteDescriptor
	Handler          any
}

func NewService(name string, connection Connection, handler any) *Service {
	return &Service{
		Name:       name,
		Connection: connection,
		Handler:    handler,
	}
}

func (s *Service) Start() error {
	if s.Connection == nil {
		return fmt.Errorf(
			"cannot start local service. The associated gateway already handles " +
				"incoming requests",
		)
	}

	if len(s.GatewayNames) == 0 {
		s.GatewayNames = []string{"*"}
	}

	for _, gatewayName := range s.GatewayNames {
		err := s.Connection.BindGatewayAnnounce(gatewayName, func(serviceDescriptors []*ServiceDescriptor) {
			s.handleGatewayAnnounce(gatewayName, serviceDescriptors)
		})
		if err != nil {
			return err
		}
	}

	err := s.Connection.BindDispatch(s.Name, func(res http.ResponseWriter, req *http.Request) {
		ctx := navaros.NewContextWithHandler(res, req, s.Handler)
		ctx.Next()
		ctx.Finalize()
	})
	if err != nil {
		return err
	}

	for _, gatewayName := range s.GatewayNames {
		if err := s.doAnnounce(gatewayName); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) Stop() {
	for _, gatewayName := range s.GatewayNames {
		s.Connection.UnbindGatewayAnnounce(gatewayName)
	}
	s.Connection.UnbindDispatch(s.Name)
}

func (s *Service) handleGatewayAnnounce(gatewayName string, serviceDescriptors []*ServiceDescriptor) {
	foundSelf := false
	for _, descriptor := range serviceDescriptors {
		if descriptor.Name == s.Name {
			foundSelf = true
			break
		}
	}
	if !foundSelf {
		if err := s.doAnnounce(gatewayName); err != nil {
			panic(err)
		}
	}
}

func (s *Service) doAnnounce(gatewayName string) error {
	routeDescriptors := s.RouteDescriptors

	if routeDescriptors == nil {
		if h, ok := s.Handler.(navaros.RouterHandler); ok {
			routeDescriptors = h.RouteDescriptors()
		}
	}

	return s.Connection.AnnounceService(gatewayName, &ServiceDescriptor{
		Name:             s.Name,
		RouteDescriptors: routeDescriptors,
	})
}
