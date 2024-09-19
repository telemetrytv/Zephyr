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
	gsi        *GatewayServiceIndexer
}

func NewGateway(name string, connection Connection) *Gateway {
	return &Gateway{
		Name:       name,
		Connection: connection,
	}
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

func (g *Gateway) Stop() error {
	g.gsi = nil
	g.Connection.UnbindServiceAnnounce(g.Name)
	return nil
}

func (g *Gateway) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	method := navaros.HTTPMethodFromString(req.Method)
	path := req.URL.Path

	serviceName, ok := g.gsi.ResolveService(method, path)
	if !ok {
		res.WriteHeader(404)
		return
	}

	if err := g.Connection.Dispatch(serviceName, res, req); err != nil {
		panic(err)
	}
}

func (g *Gateway) Handle(ctx *navaros.Context) {
	method := ctx.Method()
	path := ctx.Path()

	serviceName, ok := g.gsi.ResolveService(method, path)
	if !ok {
		ctx.Next()
		return
	}

	if err := g.Connection.Dispatch(serviceName, ctx.ResponseWriter(), ctx.Request()); err != nil {
		panic(err)
	}
}
