package zephyr

import (
	"fmt"

	"github.com/RobertWHurst/navaros"
)

type Service struct {
	GatewayNames []string
	Name         string
	Connection   Connection
	Handler      navaros.RouterHandler
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

	err := s.Connection.BindDispatchFromGatewayOrService(s.Name, func(ctx *navaros.Context) {
		s.Handler.Handle(ctx)
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
	s.Connection.UnbindGatewayAnnounce()
	s.Connection.UnBindDispatchFromGatewayOrService()
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
	return s.Connection.AnnounceService(gatewayName, &ServiceDescriptor{
		Name:             s.Name,
		RouteDescriptors: s.Handler.RouteDescriptors(),
	})
}
