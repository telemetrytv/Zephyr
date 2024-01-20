package nats

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/RobertWHurst/navaros"
	"github.com/nats-io/nats.go"
)

func (c *Connection) BindDispatchFromGatewayOrService(serviceName string, handler func(ctx *navaros.Context)) error {
	sub, err := c.Connection.Subscribe(namespace("service", serviceName), func(requestMsg *nats.Msg) {
		requestMessage := &RequestMessage{}
		if err := json.Unmarshal(requestMsg.Data, requestMessage); err != nil {
			panic(err)
		}
		requestReader := c.createRequestReaderForSubject(requestMessage.ResponseSubject)
		request, err := http.NewRequest(requestMessage.Method, requestMessage.URL, requestReader)
		if err != nil {
			panic(err)
		}

		requestBodyStreamSubject := nats.NewInbox()
		requestAckMessage := &RequestAckMessage{
			BodyStreamSubject: requestBodyStreamSubject,
		}
		requestAckMessageBuf, err := json.Marshal(requestAckMessage)
		if err != nil {
			panic(err)
		}
		if err := c.Connection.Publish(requestMsg.Reply, requestAckMessageBuf); err != nil {
			panic(err)
		}

		responseWriter := c.createResponseWriterForSubject(
			requestMessage.ResponseSubject,
			requestMessage.ResponseBodyStreamSubject,
		)

		ctx := navaros.NewContextWithHandler(responseWriter, request, handler)

		ctx.Next()
		ctx.Finalize()
		responseWriter.End()
	})

	if err != nil {
		return err
	}
	c.onUnBindDispatchFromGatewayOrService = func() {
		sub.Unsubscribe()
	}
	return nil
}

func (c *Connection) UnbindGatewayAnnounce() {
	if c.onUnbindAnnounceGateway != nil {
		c.onUnbindAnnounceGateway()
	}
}

func (c *Connection) UnBindDispatchFromGatewayOrService() {
	if c.onUnBindDispatchFromGatewayOrService != nil {
		c.onUnBindDispatchFromGatewayOrService()
	}
}

type RequestReader struct {
	Subscription *nats.Subscription
	Buffer       *bytes.Buffer
	EOF          bool
}

func (r *RequestReader) Read(p []byte) (int, error) {
	if r.EOF {
		return 0, io.EOF
	}

	if r.Buffer.Len() > 0 {
		lenRead, err := r.Buffer.Read(p)
		if err != nil {
			return 0, err
		}

		if r.Buffer.Len() >= len(p) {
			r.Buffer = bytes.NewBuffer(r.Buffer.Bytes()[lenRead:])
		} else {
			r.Buffer.Reset()
		}

		return lenRead, nil
	}

	msg, err := r.Subscription.NextMsg(DispatchTimeout)
	if err != nil {
		// if err == nats.ErrTimeout {
		// TODO: perhaps retry?
		// }
		return 0, err
	}

	bodyChunk := &BodyChunkMessage{}
	if err := json.Unmarshal(msg.Data, bodyChunk); err != nil {
		return 0, err
	}
	if bodyChunk.EOF {
		r.EOF = true
		return 0, io.EOF
	}

	// TODO: verify the index order is correct

	lenCopied := copy(p, bodyChunk.Data)
	if lenCopied < len(bodyChunk.Data) {
		r.Buffer.Write(bodyChunk.Data[lenCopied:])
	}

	return lenCopied, nil
}

func (c *Connection) createRequestReaderForSubject(subject string) io.Reader {
	sub, err := c.Connection.SubscribeSync(subject)
	if err != nil {
		panic(err)
	}

	return &RequestReader{
		Subscription: sub,
		Buffer:       &bytes.Buffer{},
	}
}

type ResponseWriter struct {
	Connection                *Connection
	headerMap                 http.Header
	responseSubject           string
	responseBodyStreamSubject string
	hasWrittenHead            bool
	writeIndex                int
}

func (r *ResponseWriter) Header() http.Header {
	return r.headerMap
}

func (r *ResponseWriter) WriteHeader(statusCode int) {
	if r.hasWrittenHead {
		panic("cannot write headers twice")
	}
	r.hasWrittenHead = true

	headers := map[string]string{}
	for key, values := range r.headerMap {
		headers[key] = values[0]
	}
	responseMessage := &ResponseMessage{
		StatusCode: statusCode,
		Headers:    headers,
	}
	responseMessageBuf, err := json.Marshal(responseMessage)
	if err != nil {
		panic(err)
	}

	if err := r.Connection.Connection.Publish(r.responseSubject, responseMessageBuf); err != nil {
		panic(err)
	}
}

func (r *ResponseWriter) Write(data []byte) (int, error) {
	if !r.hasWrittenHead {
		r.WriteHeader(200)
	}

	bodyChunk := &BodyChunkMessage{
		Index: r.writeIndex,
		Data:  data,
	}
	bodyChunkBuf, err := json.Marshal(bodyChunk)
	if err != nil {
		panic(err)
	}

	if err := r.Connection.Connection.Publish(r.responseBodyStreamSubject, bodyChunkBuf); err != nil {
		panic(err)
	}

	r.writeIndex += 1

	return len(data), nil
}

func (r *ResponseWriter) End() {
	bodyChunk := &BodyChunkMessage{
		Index: r.writeIndex,
		EOF:   true,
	}
	bodyChunkBuf, err := json.Marshal(bodyChunk)
	if err != nil {
		panic(err)
	}
	if err := r.Connection.Connection.Publish(r.responseBodyStreamSubject, bodyChunkBuf); err != nil {
		panic(err)
	}
}

func (c *Connection) createResponseWriterForSubject(resSubject, resBodySubject string) *ResponseWriter {
	return &ResponseWriter{
		Connection:                c,
		headerMap:                 http.Header{},
		responseSubject:           resSubject,
		responseBodyStreamSubject: resBodySubject,
		hasWrittenHead:            false,
		writeIndex:                0,
	}
}
