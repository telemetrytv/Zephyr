package nats

import (
	"encoding/json"
	"io"

	"github.com/RobertWHurst/navaros"
	"github.com/nats-io/nats.go"
)

func (c *Connection) DispatchToService(serviceName string, ctx *navaros.Context) error {
	responseSubject := nats.NewInbox()
	responseBodyStreamSubject := nats.NewInbox()

	requestMessage := &RequestMessage{
		Method:                    string(ctx.Method()),
		URL:                       ctx.URL().String(),
		Params:                    map[string]string(ctx.Params()),
		Headers:                   map[string][]string(ctx.RequestHeaders()),
		ResponseSubject:           responseSubject,
		ResponseBodyStreamSubject: responseBodyStreamSubject,
	}
	requestMessageBuf, err := json.Marshal(requestMessage)
	if err != nil {
		return err
	}

	responseSub, err := c.Connection.SubscribeSync(responseSubject)
	if err != nil {
		return err
	}

	responseBodyStreamSub, err := c.Connection.SubscribeSync(responseBodyStreamSubject)
	if err != nil {
		return err
	}

	requestAckMsg, err := c.Connection.Request(namespace("service", serviceName), requestMessageBuf, DispatchTimeout)
	if err != nil {
		// if err == nats.ErrTimeout {
		// TODO: perhaps retry?
		// }
		return err
	}

	requestAckMessage := &RequestAckMessage{}
	if err := json.Unmarshal(requestAckMsg.Data, requestAckMessage); err != nil {
		return err
	}

	if err := c.sendRequestBody(requestAckMessage.BodyStreamSubject, ctx); err != nil {
		return err
	}

	responseMsg, err := responseSub.NextMsg(DispatchTimeout)
	if err != nil {
		// if err == nats.ErrTimeout {
		// TODO: perhaps retry?
		// }
		return err
	}
	responseMessage := &ResponseMessage{}
	if err := json.Unmarshal(responseMsg.Data, responseMessage); err != nil {
		return err
	}

	if err := c.receiveResponseBody(responseBodyStreamSub, ctx); err != nil {
		return err
	}

	return nil
}

func (c *Connection) sendRequestBody(bodyStreamSubject string, ctx *navaros.Context) error {
	reqBodyReader := ctx.RequestBodyReader()

	for i := 0; true; i += 1 {
		buf := make([]byte, 1024)
		lenRead, err := reqBodyReader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		isEOF := err == io.EOF
		bodyChunk := &BodyChunkMessage{
			Index: i,
		}

		if isEOF {
			bodyChunk.EOF = true
		} else {
			bodyChunk.Data = buf[:lenRead]
		}

		bodyChunkBuf, err := json.Marshal(bodyChunk)
		if err != nil {
			return err
		}

		if err := c.Connection.Publish(bodyStreamSubject, bodyChunkBuf); err != nil {
			return err
		}

		if isEOF {
			break
		}
	}

	return nil
}

func (c *Connection) receiveResponseBody(sub *nats.Subscription, ctx *navaros.Context) error {
	for {

		msg, err := sub.NextMsg(DispatchTimeout)
		if err != nil {
			// if err == nats.ErrTimeout {
			// TODO: perhaps retry?
			// }
			return err
		}

		bodyChunk := &BodyChunkMessage{}
		if err := json.Unmarshal(msg.Data, bodyChunk); err != nil {
			panic(err)
		}

		if bodyChunk.EOF {
			break
		}

		// TODO: check the index and ensure we are writing the correct chunk
		if _, err := ctx.Write(bodyChunk.Data); err != nil {
			panic(err)
		}

	}

	return nil
}
