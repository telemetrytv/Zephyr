package main

import (
	"github.com/RobertWHurst/navaros"
	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
	natsconnection "github.com/telemetrytv/zephyr/nats-connection"
)

func main() {
	natsConn, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		panic(err)
	}
	conn := natsconnection.NewConnection(natsConn)

	router := &navaros.Router{}

	client := zephyr.Client{
		Connection: conn,
	}

	router.PublicGet("/leap", client.Service("example-service-a"))

	service := zephyr.Service{
		Name:         "example-service-b",
		GatewayNames: []string{"example-gateway"},
		Connection:   conn,
		Handler:      router,
	}
	if err := service.Start(); err != nil {
		panic(err)
	}

	channel := make(chan struct{})
	<-channel
}
