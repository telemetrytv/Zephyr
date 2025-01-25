package main

import (
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
	natstransport "github.com/telemetrytv/zephyr/nats-transport"
)

func main() {
	natsConn, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		panic(err)
	}

	gateway := zephyr.NewGateway("example-gateway", natstransport.New(natsConn))
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: gateway,
	}

	if err := gateway.Start(); err != nil {
		panic(err)
	}

	if err = httpServer.ListenAndServe(); err != nil {
		panic(err)
	}
}
