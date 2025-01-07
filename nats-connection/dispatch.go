package natsconnection

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/nats-io/nats.go"
)

const DispatchTimeout = 30 * time.Second
const DispatchBodyChunkSize = 1024 * 16

// TODO: look into messagepack and protobufs for more efficient serialization
// Marcus mentioned something called flatbuffers from the gaming industry

type TLS struct {
	Version            uint16 `json:"version"`
	HandshakeComplete  bool   `json:"handshakeComplete"`
	DidResume          bool   `json:"didResume"`
	CipherSuite        uint16 `json:"cipherSuite"`
	NegotiatedProtocol string `json:"negotiatedProtocol"`
	ServerName         string `json:"serverName"`
	// NOTE: Due to the complexity of the certificate structures, we are not
	//       including them in the JSON output for now.
	// PeerCertificates            []JSONCertificate   `json:"peerCertificates"`
	// VerifiedChains              [][]JSONCertificate `json:"verifiedChains"`
	SignedCertificateTimestamps [][]byte `json:"signedCertificateTimestamps"`
	OCSPResponse                []byte   `json:"ocspResponse"`
	TLSUnique                   []byte   `json:"tlsUnique"`
}

type Request struct {
	Method           string              `json:"method"`
	URL              string              `json:"url"`
	Proto            string              `json:"proto"`
	ProtoMajor       int                 `json:"protoMajor"`
	ProtoMinor       int                 `json:"protoMinor"`
	Header           map[string][]string `json:"header"`
	ContentLength    int64               `json:"contentLength"`
	TransferEncoding []string            `json:"transferEncoding"`
	Host             string              `json:"host"`
	Trailers         map[string][]string `json:"trailers"`
	RemoteAddr       string              `json:"remoteAddr"`
	RequestURI       string              `json:"requestURI"`
	TLS              *TLS                `json:"tls"`

	ResponseSubject     string `json:"responseSubject"`
	ResponseBodySubject string `json:"responseBodySubject"`
}

type RequestAck struct {
	RequestBodySubject string `json:"requestBodySubject"`
}

type ResponseError struct {
	Message string `json:"message"`
	Stack   string `json:"stack"`
}

type Response struct {
	StatusCode int                 `json:"statusCode"`
	Header     map[string][]string `json:"header"`
	Error      string              `json:"error"`
}

type BodyChunk struct {
	Index int    `json:"index"`
	Data  []byte `json:"data"`
	Error string `json:"error"`
	IsEOF bool   `json:"end"`
}

func (c *Connection) Dispatch(serviceName string, res http.ResponseWriter, req *http.Request) error {
	requestSubject := namespace("service", serviceName)
	responseSubject := nats.NewInbox()
	responseBodySubject := nats.NewInbox()

	request := &Request{
		Method:           req.Method,
		URL:              req.URL.String(),
		Proto:            req.Proto,
		ProtoMajor:       req.ProtoMajor,
		ProtoMinor:       req.ProtoMinor,
		Header:           req.Header,
		ContentLength:    req.ContentLength,
		TransferEncoding: req.TransferEncoding,
		Host:             req.Host,
		Trailers:         req.Trailer,
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       req.RequestURI,

		ResponseSubject:     responseSubject,
		ResponseBodySubject: responseBodySubject,
	}

	if req.TLS != nil {
		request.TLS = &TLS{
			Version:                     req.TLS.Version,
			HandshakeComplete:           req.TLS.HandshakeComplete,
			DidResume:                   req.TLS.DidResume,
			CipherSuite:                 req.TLS.CipherSuite,
			NegotiatedProtocol:          req.TLS.NegotiatedProtocol,
			ServerName:                  req.TLS.ServerName,
			SignedCertificateTimestamps: req.TLS.SignedCertificateTimestamps,
			OCSPResponse:                req.TLS.OCSPResponse,
			TLSUnique:                   req.TLS.TLSUnique,
		}
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	responseSub, err := c.NatsConnection.SubscribeSync(responseSubject)
	if err != nil {
		return err
	}
	responseBodySub, err := c.NatsConnection.SubscribeSync(responseBodySubject)
	if err != nil {
		return err
	}

	requestAckMsg, err := c.NatsConnection.Request(requestSubject, requestBytes, DispatchTimeout)
	if err != nil {
		return err
	}
	requestAck := &RequestAck{}
	if err := json.Unmarshal(requestAckMsg.Data, requestAck); err != nil {
		return err
	}

	requestBodySubject := requestAck.RequestBodySubject

	for i := 0; true; i += 1 {
		requestBodyBytes := make([]byte, DispatchBodyChunkSize)
		lenRead, err := req.Body.Read(requestBodyBytes)
		isEOF := err == io.EOF
		if !isEOF && err != nil {
			return err
		}

		bodyChunk := &BodyChunk{
			Index: i,
			IsEOF: isEOF,
		}
		if !isEOF {
			bodyChunk.Data = requestBodyBytes[:lenRead]
		}
		bodyChunkBytes, err := json.Marshal(bodyChunk)
		if err != nil {
			return err
		}

		println("sending body chunk", i)
		if err := c.NatsConnection.Publish(requestBodySubject, bodyChunkBytes); err != nil {
			return err
		}

		if isEOF {
			break
		}
	}

	responseMsg, err := responseSub.NextMsg(DispatchTimeout)
	if err != nil {
		return err
	}
	if err := responseSub.Unsubscribe(); err != nil {
		return err
	}

	response := &Response{}
	if err := json.Unmarshal(responseMsg.Data, response); err != nil {
		return err
	}

	for key, values := range response.Header {
		for _, value := range values {
			res.Header().Add(key, value)
		}
	}
	res.WriteHeader(response.StatusCode)

	for i := 0; true; i += 1 {
		bodyChunkMsg, err := responseBodySub.NextMsg(DispatchTimeout)
		if err != nil {
			return err
		}

		bodyChunk := &BodyChunk{}
		if err := json.Unmarshal(bodyChunkMsg.Data, bodyChunk); err != nil {
			return err
		}

		if _, err := res.Write(bodyChunk.Data); err != nil {
			return err
		}

		if bodyChunk.IsEOF {
			break
		}
	}

	if err := responseBodySub.Unsubscribe(); err != nil {
		return err
	}

	return nil
}

func (c *Connection) BindDispatch(serviceName string, handler func(res http.ResponseWriter, req *http.Request)) error {
	dispatchSubject := namespace("service", serviceName)
	sub, err := c.NatsConnection.QueueSubscribe(dispatchSubject, dispatchSubject, func(msg *nats.Msg) {
		if err := c.handleDispatch(msg, handler); err != nil {
			panic(err)
		}
	})
	if err != nil {
		return err
	}

	unbinders, ok := c.unbindDispatch[serviceName]
	if !ok {
		unbinders = []func(){}
	}
	unbinders = append(unbinders, func() {
		if err := sub.Unsubscribe(); err != nil {
			panic(err)
		}
	})
	c.unbindDispatch[serviceName] = unbinders

	return nil
}

type requestReader struct {
	natsSubscription *nats.Subscription
	hasEnded         bool
	buffer           bytes.Buffer
}

func (r *requestReader) Read(p []byte) (int, error) {
	if !r.hasEnded {
		readCount := int(math.Ceil(float64(len(p)-r.buffer.Len()) / DispatchBodyChunkSize))

		// TODO: should keep track of the index
		for i := 0; i < readCount; i += 1 {
			bodyChunkMsg, err := r.natsSubscription.NextMsg(DispatchTimeout)
			if err != nil {
				return 0, err
			}
			bodyChunk := &BodyChunk{}
			if err := json.Unmarshal(bodyChunkMsg.Data, bodyChunk); err != nil {
				return 0, err
			}

			if bodyChunk.IsEOF {
				r.hasEnded = true
				break
			}

			if _, err := r.buffer.Write(bodyChunk.Data); err != nil {
				return 0, err
			}
		}

		if r.hasEnded {
			if err := r.natsSubscription.Unsubscribe(); err != nil {
				return 0, err
			}
		}
	}

	return r.buffer.Read(p)
}

func (r *requestReader) Close() error {
	// TODO: currently is a noop, in future we may want to unsubscribe and
	// inform the other side that we are closing the connection
	return nil
}

type responseWriter struct {
	responseSubject     string
	responseBodySubject string
	natsConnection      *nats.Conn
	header              http.Header
	statusCode          int
	hasSentHeaders      bool
	writeIndex          int
	buffer              bytes.Buffer
	err                 error
}

func (r *responseWriter) Header() http.Header {
	return r.header
}

func (r *responseWriter) WriteHeader(statusCode int) {
	if r.statusCode != 0 {
		panic("cannot write headers twice")
	}
	r.statusCode = statusCode
}

func (r *responseWriter) Write(p []byte) (int, error) {
	if err := r.ensureHeadersSent(); err != nil {
		return 0, err
	}
	len, err := r.buffer.Write(p)
	if err != nil {
		return 0, err
	}
	chunkCount := int(math.Floor(float64(r.buffer.Len()) / DispatchBodyChunkSize))
	for i := 0; i < chunkCount; i += 1 {
		if err := r.writeChunk(); err != nil {
			return 0, err
		}
	}
	return len, nil
}

func (r *responseWriter) WriteError(err error) {
	r.err = err
}

func (r *responseWriter) End() error {
	if err := r.ensureHeadersSent(); err != nil {
		return err
	}
	if r.buffer.Len() > 0 {
		if err := r.writeChunk(); err != nil {
			return err
		}
	}

	bodyChunk := &BodyChunk{
		Index: r.writeIndex,
		IsEOF: true,
	}
	if r.err != nil {
		bodyChunk.Error = r.err.Error()
	}
	bodyChunkBytes, err := json.Marshal(bodyChunk)
	if err != nil {
		return err
	}

	println("writing end", r.writeIndex)
	return r.natsConnection.Publish(r.responseBodySubject, bodyChunkBytes)
}

func (r *responseWriter) ensureHeadersSent() error {
	if r.hasSentHeaders {
		return nil
	}
	r.hasSentHeaders = true

	headers := map[string][]string{}
	for key, values := range r.header {
		headers[key] = values
	}
	response := &Response{
		StatusCode: r.statusCode,
		Header:     headers,
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return r.natsConnection.Publish(r.responseSubject, responseBytes)
}

func (r *responseWriter) writeChunk() error {
	bodyChunkBytes, err := json.Marshal(&BodyChunk{
		Index: r.writeIndex,
		Data:  r.buffer.Next(DispatchBodyChunkSize),
	})
	if err != nil {
		return err
	}
	println("writing body chunk", r.writeIndex)
	if err := r.natsConnection.Publish(r.responseBodySubject, bodyChunkBytes); err != nil {
		return err
	}
	r.writeIndex += 1
	return nil
}

func (c *Connection) handleDispatch(msg *nats.Msg, handler func(res http.ResponseWriter, req *http.Request)) error {
	responseBodySubject := nats.NewInbox()

	request := &Request{}
	if err := json.Unmarshal(msg.Data, request); err != nil {
		return err
	}

	reqUrl, err := url.Parse(request.URL)
	if err != nil {
		return err
	}

	reqBodySubscription, err := c.NatsConnection.SubscribeSync(responseBodySubject)
	if err != nil {
		return err
	}

	reqReader := &requestReader{
		natsSubscription: reqBodySubscription,
		buffer:           bytes.Buffer{},
	}

	res := &responseWriter{
		responseSubject:     request.ResponseSubject,
		responseBodySubject: request.ResponseBodySubject,
		natsConnection:      c.NatsConnection,
		header:              map[string][]string{},
		buffer:              bytes.Buffer{},
	}

	req := &http.Request{
		Method:           request.Method,
		URL:              reqUrl,
		Proto:            request.Proto,
		ProtoMajor:       request.ProtoMajor,
		ProtoMinor:       request.ProtoMinor,
		Header:           request.Header,
		ContentLength:    request.ContentLength,
		TransferEncoding: request.TransferEncoding,
		Host:             request.Host,
		Trailer:          request.Trailers,
		RemoteAddr:       request.RemoteAddr,
		RequestURI:       request.RequestURI,
		Body:             reqReader,
	}

	if request.TLS != nil {
		req.TLS = &tls.ConnectionState{
			Version:                     request.TLS.Version,
			HandshakeComplete:           request.TLS.HandshakeComplete,
			DidResume:                   request.TLS.DidResume,
			CipherSuite:                 request.TLS.CipherSuite,
			NegotiatedProtocol:          request.TLS.NegotiatedProtocol,
			ServerName:                  request.TLS.ServerName,
			SignedCertificateTimestamps: request.TLS.SignedCertificateTimestamps,
			OCSPResponse:                request.TLS.OCSPResponse,
			TLSUnique:                   request.TLS.TLSUnique,
		}
	}

	requestAck := &RequestAck{
		RequestBodySubject: responseBodySubject,
	}
	requestAckBytes, err := json.Marshal(requestAck)
	if err != nil {
		return err
	}
	if err := c.NatsConnection.Publish(msg.Reply, requestAckBytes); err != nil {
		return err
	}

	func() {
		defer func() {
			if err := recover(); err != nil {
				res.WriteError(err.(error))
			}
		}()

		handler(res, req)
	}()

	if err := res.End(); err != nil {
		return err
	}

	return nil
}

func (c *Connection) UnbindDispatch(serviceName string) {
	unbinders, ok := c.unbindDispatch[serviceName]
	if ok {
		for _, unbind := range unbinders {
			unbind()
		}
	}
}
