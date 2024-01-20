package main

import (
	"encoding/json"
	"io"

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

	router := &navaros.Router{}

	router.Get("/hello-world", func(ctx *navaros.Context) {
		resBody := ""
		for i := 0; i < 10000; i++ {
			resBody += "Hello, world! "
		}
		ctx.Body = resBody
	})

	router.Post("/hello-world", func(ctx *navaros.Context) {
		reqBody, err := io.ReadAll(ctx.RequestBodyReader())
		if err != nil {
			panic(err)
		}
		ctx.Body = reqBody
	})

	router.Get("/request-info", func(ctx *navaros.Context) {
		body := map[string]any{
			"method":                  ctx.Method(),
			"path":                    ctx.Path(),
			"url":                     ctx.URL().String(),
			"params":                  ctx.Params(),
			"query":                   ctx.Query(),
			"protocol":                ctx.RequestProtocol(),
			"protocolMajor":           ctx.RequestProtocolMajor(),
			"protocolMinor":           ctx.RequestProtocolMinor(),
			"headers":                 ctx.RequestHeaders(),
			"contentLength":           ctx.RequestContentLength(),
			"requestTransferEncoding": ctx.RequestTransferEncoding(),
			"requestHost":             ctx.RequestHost(),
			"requestRemoteAddress":    ctx.RequestRemoteAddress(),
			"requestRawURI":           ctx.RequestRawURI(),
			"requestTLS":              ctx.RequestTLS(),
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}

		ctx.Headers.Add("Content-Type", "application/json")
		ctx.Body = bodyBytes
	})

	service := zephyr.Service{
		Name:         "example-service",
		GatewayNames: []string{"example-gateway"},
		Connection:   natsconnection.NewConnection(natsConn),
		Handler:      router,
	}
	if err := service.Start(); err != nil {
		panic(err)
	}

	channel := make(chan struct{})
	<-channel
}
