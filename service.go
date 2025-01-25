package zephyr

import (
	"fmt"
	"net/http"

	"github.com/RobertWHurst/navaros"
)

// Service is a struct that facilitates communication between a go microservice
// and a zephyr gateway. It will manage the announcement of the service to the
// gateway as well as calls any HTTP Handler or Navaros handler.
//
// If the service is given a Navaros router, it will automatically announce
// any public routes declared on the router. That said any handler compatible
// with go's http.HandlerFunc or http.Handler interface can be used.
//
// Note that if you opt to use something other than Navaros, you will need to
// assign your route descriptors manually. This can be done by using the
// zephyr.NewRouteDescriptor function to create your route descriptors, then
// assigning them to the RouteDescriptors field on the service.
type Service struct {

	// GatewayNames is a list of gateway names that the service should announce
	// itself to. If the list is empty, the service will announce itself to all
	// gateways on the connection.
	GatewayNames []string

	// Name is the name of the service. This is used to identify the service
	// when announcing it to the gateway.
	Name string

	// Transport is a struct that implements the Transport interface and
	// facilitates communication between services, gateways, and clients.
	Transport Transport

	// RouteDescriptors is a list of route descriptors that describe the routes
	// that this service can handle. If this is left empty, the service will
	// not be routable. This is automatically populated if the handler is a
	// Navaros router.
	RouteDescriptors []*navaros.RouteDescriptor

	// Handler is called when a request is made to the service. This can be
	// either a Navaros router or a standard http.Handler or http.HandlerFunc.
	Handler any

	stopChan chan struct{}
}

// NewService creates a new service with the given name, connection, and handler.
// The service will automatically announce itself to the gateway when it starts.
// If the handler is a Navaros router, the service will automatically announce
// any public routes declared on the router.
func NewService(name string, transport Transport, handler any) *Service {
	return &Service{
		Name:      name,
		Transport: transport,
		Handler:   handler,
		stopChan:  make(chan struct{}),
	}
}

// Start starts the service. This will bind the service to the connection and
// announce the service to the gateway. If the service is already bound to the
// connection, this will return an error. Start must be called before the
// service will receive requests.
func (s *Service) Start() error {
	if s.Transport == nil {
		return fmt.Errorf(
			"cannot start local service. The associated gateway already handles " +
				"incoming requests",
		)
	}

	err := s.Transport.BindGatewayAnnounce(func(gatewayDescriptor *GatewayDescriptor) {
		s.handleGatewayAnnounce(gatewayDescriptor)
	})
	if err != nil {
		return err
	}

	err = s.Transport.BindDispatch(s.Name, func(res http.ResponseWriter, req *http.Request) {
		ctx := navaros.NewContext(res, req, s.Handler)
		ctx.Next()
		navaros.CtxFinalize(ctx)
	})
	if err != nil {
		return err
	}

	if err := s.doAnnounce(); err != nil {
		return err
	}

	return nil
}

// Stop stops the service. This will unbind the service from the connection.
// This provides a way to dispose of the service if need be.
func (s *Service) Stop() {
	if err := s.Transport.UnbindGatewayAnnounce(); err != nil {
		panic(err)
	}
	if err := s.Transport.UnbindDispatch(s.Name); err != nil {
		panic(err)
	}
	close(s.stopChan)
}

func (s *Service) handleGatewayAnnounce(GatewayDescriptor *GatewayDescriptor) {
	isWantedGateway := len(s.GatewayNames) == 0
	if !isWantedGateway {
		for _, name := range s.GatewayNames {
			if name == GatewayDescriptor.Name {
				isWantedGateway = true
				break
			}
		}
	}
	if !isWantedGateway {
		return
	}

	foundSelf := false
	for _, descriptor := range GatewayDescriptor.ServiceDescriptors {
		if descriptor.Name == s.Name {
			foundSelf = true
			break
		}
	}
	if !foundSelf {
		if err := s.doAnnounce(); err != nil {
			panic(err)
		}
	}
}

// Run starts the service and blocks until the service is stopped.
func (s *Service) Run() error {
	if err := s.Start(); err != nil {
		return err
	}
	<-s.stopChan
	return nil
}

func (s *Service) doAnnounce() error {
	routeDescriptors := s.RouteDescriptors

	if routeDescriptors == nil {
		if h, ok := s.Handler.(navaros.RouterHandler); ok {
			routeDescriptors = h.RouteDescriptors()
		}
	}

	return s.Transport.AnnounceService(&ServiceDescriptor{
		Name:             s.Name,
		GatewayNames:     s.GatewayNames,
		RouteDescriptors: routeDescriptors,
	})
}
