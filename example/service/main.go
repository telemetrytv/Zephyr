package main

import (
	"github.com/RobertWHurst/navaros"
	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
	zephyrNats "github.com/telemetrytv/zephyr/nats"
)

func main() {
	natsConn, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		panic(err)
	}

	router := &navaros.Router{}

	router.Get("/hello-world", func(ctx *navaros.Context) {
		resBody := ""
		for i := 0; i < 10000; i++ {
			resBody += "Hello, world! "
		}
		ctx.Body = resBody
	})

	service := zephyr.Service{
		Name:         "example-service",
		GatewayNames: []string{"example-gateway"},
		Connection:   zephyrNats.NewConnection(natsConn),
		Handler:      router,
	}

	if err := service.Start(); err != nil {
		panic(err)
	}

	channel := make(chan struct{})
	<-channel
}
