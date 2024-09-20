package zephyr

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/RobertWHurst/navaros"
	"github.com/coder/websocket"
)

var GatewayAnnounceInterval = time.Duration((8 + rand.Intn(2))) * time.Second

type Gateway struct {
	Name                   string
	Connection             Connection
	AllowWebSockets        bool
	AcceptWebSocketOrigins []string
	gsi                    *GatewayServiceIndexer
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

func (g *Gateway) Stop() {
	g.gsi = nil
	g.Connection.UnbindServiceAnnounce(g.Name)
}

func (g *Gateway) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	serviceName, ok := g.gsi.ResolveService(navaros.HTTPMethodFromString(req.Method), req.URL.Path)
	if !ok {
		res.WriteHeader(404)
		return
	}

	if g.tryUpgradeIfWebSocketRequest(res, req) {
		return
	}

	if err := g.Connection.Dispatch(serviceName, res, req); err != nil {
		panic(err)
	}
}

func (g *Gateway) Handle(ctx *navaros.Context) {
	serviceName, ok := g.gsi.ResolveService(ctx.Method(), ctx.Path())
	if !ok {
		ctx.Next()
		return
	}

	if g.tryUpgradeIfWebSocketRequest(ctx.ResponseWriter(), ctx.Request()) {
		return
	}

	if err := g.Connection.Dispatch(serviceName, ctx.ResponseWriter(), ctx.Request()); err != nil {
		panic(err)
	}
}

func (g *Gateway) CanServeHTTP(req *http.Request) bool {
	method := navaros.HTTPMethodFromString(req.Method)
	path := req.URL.Path
	_, ok := g.gsi.ResolveService(method, path)
	return ok
}

func (g *Gateway) CanHandle(ctx *navaros.Context) bool {
	method := ctx.Method()
	path := ctx.Path()
	_, ok := g.gsi.ResolveService(method, path)
	return ok
}

func (g *Gateway) tryUpgradeIfWebSocketRequest(res http.ResponseWriter, req *http.Request) bool {
	if !g.AllowWebSockets {
		return false
	}

	if req.Header.Get("Connection") != "Upgrade" {
		return false
	}

	method := navaros.HTTPMethodFromString(req.Method)
	path := req.URL.Path
	serviceName, ok := g.gsi.ResolveService(method, path)
	if !ok {
		return false
	}

	conn, err := websocket.Accept(res, req, &websocket.AcceptOptions{
		OriginPatterns: g.AcceptWebSocketOrigins,
	})
	if err != nil {
		panic(err)
	}
	if err := g.Connection.AttachSocketConnection(serviceName, conn); err != nil {
		panic(err)
	}

	return true
}
