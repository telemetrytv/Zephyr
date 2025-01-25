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
	Name      string
	Transport Transport
	gsi       *GatewayServiceIndexer
}

var _ http.Handler = &Gateway{}
var _ navaros.Handler = &Gateway{}

func NewGateway(name string, transport Transport) *Gateway {
	return &Gateway{
		Name:      name,
		Transport: transport,
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

	if g.Transport == nil {
		return fmt.Errorf(
			"cannot start local service. The associated gateway already handles " +
				"incoming requests",
		)
	}

	err := g.Transport.BindServiceAnnounce(func(serviceDescriptor *ServiceDescriptor) {
		announcingToThisGateway := len(serviceDescriptor.GatewayNames) == 0
		if !announcingToThisGateway {
			for _, gatewayName := range serviceDescriptor.GatewayNames {
				if gatewayName == g.Name {
					announcingToThisGateway = true
					break
				}
			}
		}
		if !announcingToThisGateway {
			return
		}

		if err := g.gsi.SetServiceDescriptor(serviceDescriptor); err != nil {
			panic(err)
		}
	})
	if err != nil {
		return err
	}

	return g.Transport.AnnounceGateway(&GatewayDescriptor{
		Name:               g.Name,
		ServiceDescriptors: g.gsi.ServiceDescriptors,
	})
}

func (g *Gateway) Stop() {
	if g.gsi == nil {
		return
	}
	g.gsi = nil
	if err := g.Transport.UnbindServiceAnnounce(); err != nil {
		panic(err)
	}
}

func (g *Gateway) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if g.gsi == nil {
		res.WriteHeader(503)
		return
	}

	method, err := navaros.HTTPMethodFromString(req.Method)
	if err != nil {
		res.WriteHeader(400)
	}

	path := req.URL.Path

	serviceName, ok := g.gsi.ResolveService(method, path)
	if !ok {
		res.WriteHeader(404)
		return
	}

	if err := g.Transport.Dispatch(serviceName, res, req); err != nil {
		panic(err)
	}
}

func (g *Gateway) CanServeHTTP(req *http.Request) bool {
	if g.gsi == nil {
		return false
	}

	method, err := navaros.HTTPMethodFromString(req.Method)
	if err != nil {
		return false
	}

	path := req.URL.Path
	_, ok := g.gsi.ResolveService(method, path)
	return ok
}

func (g *Gateway) Handle(ctx *navaros.Context) {
	if g.gsi == nil {
		ctx.Next()
		return
	}

	method := ctx.Method()
	path := ctx.Path()

	serviceName, ok := g.gsi.ResolveService(method, path)
	if !ok {
		ctx.Next()
		return
	}

	// This Panic is ok because it will be caught and handled by Navaros
	if err := g.Transport.Dispatch(serviceName, ctx.ResponseWriter(), ctx.Request()); err != nil {
		panic(err)
	}
}

func (g *Gateway) CanHandle(ctx *navaros.Context) bool {
	if g.gsi == nil {
		return false
	}

	method := ctx.Method()
	path := ctx.Path()
	_, ok := g.gsi.ResolveService(method, path)
	return ok
}
