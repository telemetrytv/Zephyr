package zephyr

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/RobertWHurst/navaros"
)

var GatewayAnnounceInterval = time.Duration((8 + rand.Intn(2))) * time.Second

type Gateway struct {
	Name       string
	Connection Connection
	Router     *navaros.Router
	gsi        *GatewayServiceIndexer
}

// Announce runs a loop which sends a message to all services periodically,
// introducing them to this gateway if they are not already aware of it.
// The announcement message contains information about the services that
// this gateway is aware of. Services that do not see themselves in the
// announcement message are expected to send a reply to the gateway with their
// routing information.
func (g *Gateway) Start() error {
	if g.gsi != nil {
		return fmt.Errorf("gateway already started")
	}

	g.gsi = &GatewayServiceIndexer{
		ServiceDescriptors: []*ServiceDescriptor{},
	}

	if g.Connection == nil {
		return fmt.Errorf(
			"cannot start local service. The associated gateway already handles " +
				"incoming requests",
		)
	}

	err := g.Connection.BindServiceAnnounce(g.Name, func(serviceDescriptor *ServiceDescriptor) {
		g.gsi.SetServiceDescriptor(serviceDescriptor)
	})
	if err != nil {
		return err
	}

	return g.Connection.AnnounceGateway(g.Name, g.gsi.ServiceDescriptors)
}

func (g *Gateway) Stop() {
	g.gsi = nil
	g.Connection.UnbindServiceAnnounce()
}

// ServeHTTP implements the http.Handler interface. It is provided for convenience
// so the gateway can be used as a http.Handler. If you need more control it's
// recommended to use gateway as a navaros.Handler instead.
func (g *Gateway) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if g.Router == nil {
		g.Router = &navaros.Router{}
		// TODO: Add some decent defaults like a 404 handler
	}
	g.Router.Use(g)
	g.Router.ServeHTTP(res, req)
}

// Handle implements the navaros.Handler interface. This is the recommended way
// to use the gateway as it allows you to setup a navaros router to suit your
// needs with the gateway being just one of the handlers.
func (g *Gateway) Handle(ctx *navaros.Context) {
	if g.gsi == nil {
		panic("gateway not started")
	}

	serviceName, foundService := g.gsi.ResolveService(ctx.Method(), ctx.Path())
	if !foundService {
		ctx.Next()
		return
	}
	if err := g.Connection.DispatchToService(serviceName, ctx); err != nil {
		panic(err)
	}
}
