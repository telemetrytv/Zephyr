package natsconnection

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coder/websocket"
	"github.com/nats-io/nats.go"
)

var OpenConnectionTimeout = 5 * time.Second

type SocketState string

const (
	SocketStateOpen   SocketState = "open"
	SocketStateOpened SocketState = "opened"
	SocketStateClose  SocketState = "close"
	SocketStateClosed SocketState = "closed"
)

type SocketStateMessage struct {
	State       SocketState `json:"state"`
	RecvSubject string      `json:"recvSubject"`
}

type SocketStateAck struct {
	State       SocketState `json:"state"`
	SendSubject string      `json:"sendSubject"`
}

func (c *Connection) AttachSocketConnection(serviceName string, webSocketConnection *websocket.Conn) error {
	socketSubject := namespace("service.socket", serviceName)
	recvSubject := nats.NewInbox()

	socketOpenedMessage := SocketStateMessage{
		State:       SocketStateOpen,
		RecvSubject: recvSubject,
	}

	socketOpenedMessageBytes, err := json.Marshal(socketOpenedMessage)
	if err != nil {
		return err
	}

	openSocketResp, err := c.NatsConnection.Request(socketSubject, socketOpenedMessageBytes, OpenConnectionTimeout)
	if err != nil {
		return err
	}

	socketAck := &SocketStateAck{}
	if err := json.Unmarshal(openSocketResp.Data, socketAck); err != nil {
		return err
	}
	if socketAck.State != SocketStateOpened {
		return fmt.Errorf("socket connection not opened")
	}

	recvSub, err := c.NatsConnection.Subscribe(socketAck.SendSubject, func(msg *nats.Msg) {
		webSocketConnection.Write(context.Background(), websocket.MessageBinary, msg.Data)
	})
	if err != nil {
		return err
	}

	closeSocketChan := make(chan error)

	go func() {
		_, webSocketReader, err := webSocketConnection.Reader(context.Background())
		if err != nil {
			closeSocketChan <- err
		}

		readBuf := make([]byte, 1024)
		for {
			_, err := webSocketReader.Read(readBuf)
			if err != nil {
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
					closeSocketChan <- nil
				} else {
					closeSocketChan <- err
				}
				break
			}
			if err := c.NatsConnection.Publish(recvSubject, readBuf); err != nil {
				closeSocketChan <- err
				break
			}
		}
	}()

	err = <-closeSocketChan
	if err != nil {
		return err
	}

	socketClosedMessage := SocketStateMessage{
		State:       SocketStateClose,
		RecvSubject: recvSubject,
	}

	socketClosedMessageBytes, err := json.Marshal(socketClosedMessage)
	if err != nil {
		return err
	}

	if err := c.NatsConnection.Publish(socketSubject, socketClosedMessageBytes); err != nil {
		return err
	}
	if err := recvSub.Unsubscribe(); err != nil {
		return err
	}

	return nil
}

func (c *Connection) BindSocketConnections(serviceName string, handler func(webSocketConnection *websocket.Conn)) error {
	dispatchSubject := namespace("service.socket", serviceName)
	sub, err := c.NatsConnection.Subscribe(dispatchSubject, func(msg *nats.Msg) {
		if err := c.handleSocketConnection(msg, handler); err != nil {
			panic(err)
		}
	})
	if err != nil {
		return err
	}

	unbinders, ok := c.unbindSocketConnections[serviceName]
	if !ok {
		unbinders = []func(){}
	}
	unbinders = append(unbinders, func() { sub.Unsubscribe() })
	c.unbindSocketConnections[serviceName] = unbinders

	return nil
}

func (c *Connection) handleSocketConnection(msg *nats.Msg, handler func(webSocketConnection *websocket.Conn)) error {
	return nil
}
