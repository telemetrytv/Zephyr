# Zephyr

![Zephyr Graphic](zephyr.png)

Zephyr is a microservice framework developed by the TelemetryTV team. It contains gateway logic for streams and http as well as service registration and communication

## Setting up a REST Gateway

```go
package myservice

import (
  "net/http"
  "github.com/nats-io/nats.go"
  "github.com/telemetrytv/zephyr/service"
  "github.com/telemetrytv/zephyr/rest/router"
  "github.com/telemetrytv/zephyr/rest/gateway"
  "github.com/RobertWHurst/navaros"
)

natConn, err := nats.Connect("...")
if err != nil {
  ...
}

service := zephyr.NewService("myservice", natConn)

router := navaros.Mux{}

router.Get("/hello-world", func (ctx *Context) {
  ctx.Body = "Hello World!!!"
})

service.Handler = router

// -----


natConn, err := nats.Connect("...")
if err != nil {
  ...
}

gateway := zephyr.NewGateway(natConn)

httpServer := http.Server{
  ...
  Handler: gateway
}


```
