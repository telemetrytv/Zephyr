package main

import (
	"github.com/RobertWHurst/navaros"
	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
	natsconnection "github.com/telemetrytv/zephyr/connections/nats-connection"
)

func main() {
	natsConn, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		panic(err)
	}

	router := &navaros.Router{}

	router.Get("/leap", func(ctx *navaros.Context) {
		ctx.Body = "response from service-a via service-b"
	})

	service := zephyr.Service{
		Name:         "example-service-a",
		GatewayNames: []string{"example-gateway"},
		Connection:   natsconnection.New(natsConn),
		Handler:      router,
	}
	if err := service.Start(); err != nil {
		panic(err)
	}

	channel := make(chan struct{})
	<-channel
}
