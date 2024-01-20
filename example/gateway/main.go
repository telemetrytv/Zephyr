package main

import (
	"net/http"

	"github.com/RobertWHurst/navaros"
	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
	zephyrNats "github.com/telemetrytv/zephyr/nats"
)

func init() {
	navaros.PrintHandlerErrors = true
}

func main() {
	natsConn, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		panic(err)
	}

	gateway := &zephyr.Gateway{
		Name:       "example-gateway",
		Connection: zephyrNats.NewConnection(natsConn),
	}

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
