package natstransport

import (
	"bytes"
	"crypto/tls"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"
)

const DispatchTimeout = 30 * time.Hour
const DispatchBodyChunkSize = 1024 * 16

type TLS struct {
	Version            uint16 `msgpack:"version"`
	HandshakeComplete  bool   `msgpack:"handshakeComplete"`
	DidResume          bool   `msgpack:"didResume"`
	CipherSuite        uint16 `msgpack:"cipherSuite"`
	NegotiatedProtocol string `msgpack:"negotiatedProtocol"`
	ServerName         string `msgpack:"serverName"`
	// NOTE: Due to the complexity of the certificate structures, we are not
	//       including them in the JSON output for now.
	// PeerCertificates            []JSONCertificate   `msgpack:"peerCertificates"`
	// VerifiedChains              [][]JSONCertificate `msgpack:"verifiedChains"`
	SignedCertificateTimestamps [][]byte `msgpack:"signedCertificateTimestamps"`
	OCSPResponse                []byte   `msgpack:"ocspResponse"`
	TLSUnique                   []byte   `msgpack:"tlsUnique"`
}

type Request struct {
	Method           string              `msgpack:"method"`
	URL              string              `msgpack:"url"`
	Proto            string              `msgpack:"proto"`
	ProtoMajor       int                 `msgpack:"protoMajor"`
	ProtoMinor       int                 `msgpack:"protoMinor"`
	Header           map[string][]string `msgpack:"header"`
	ContentLength    int64               `msgpack:"contentLength"`
	TransferEncoding []string            `msgpack:"transferEncoding"`
	Host             string              `msgpack:"host"`
	Trailers         map[string][]string `msgpack:"trailers"`
	RemoteAddr       string              `msgpack:"remoteAddr"`
	RequestURI       string              `msgpack:"requestURI"`
	TLS              *TLS                `msgpack:"tls"`

	ResponseSubject     string `msgpack:"responseSubject"`
	ResponseBodySubject string `msgpack:"responseBodySubject"`
}

type RequestAck struct {
	RequestBodySubject string `msgpack:"requestBodySubject"`
}

type ResponseError struct {
	Message string `msgpack:"message"`
	Stack   string `msgpack:"stack"`
}

type Response struct {
	StatusCode int                 `msgpack:"statusCode"`
	Header     map[string][]string `msgpack:"header"`
	Error      string              `msgpack:"error"`
}

type BodyChunk struct {
	Index int    `msgpack:"index"`
	Data  []byte `msgpack:"data"`
	Error string `msgpack:"error"`
	IsEOF bool   `msgpack:"end"`
}

func (c *NatsTransport) Dispatch(serviceName string, res http.ResponseWriter, req *http.Request) error {
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

	requestBytes, err := msgpack.Marshal(request)
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
	if err := msgpack.Unmarshal(requestAckMsg.Data, requestAck); err != nil {
		return err
	}

	requestBodySubject := requestAck.RequestBodySubject

	// Cover the case of a nil body reader
	reqBody := req.Body
	if reqBody == nil {
		reqBody = &eofReader{}
	}

	i := 0
	for {
		requestBodyBytes := make([]byte, DispatchBodyChunkSize)
		lenRead, err := reqBody.Read(requestBodyBytes)
		isEOF := err == io.EOF
		if !isEOF && err != nil {
			return err
		}

		if lenRead != 0 {
			bodyChunk := &BodyChunk{
				Index: i,
				Data:  requestBodyBytes[:lenRead],
			}
			i += 1

			bodyChunkBytes, err := msgpack.Marshal(bodyChunk)
			if err != nil {
				return err
			}

			if err := c.NatsConnection.Publish(requestBodySubject, bodyChunkBytes); err != nil {
				return err
			}
		}

		if isEOF {
			bodyChunk := &BodyChunk{
				Index: i,
				IsEOF: true,
			}

			bodyChunkBytes, err := msgpack.Marshal(bodyChunk)
			if err != nil {
				return err
			}

			if err := c.NatsConnection.Publish(requestBodySubject, bodyChunkBytes); err != nil {
				return err
			}

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
	if err := msgpack.Unmarshal(responseMsg.Data, response); err != nil {
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
		if err := msgpack.Unmarshal(bodyChunkMsg.Data, bodyChunk); err != nil {
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

func (c *NatsTransport) BindDispatch(serviceName string, handler func(res http.ResponseWriter, req *http.Request)) error {
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
		unbinders = []func() error{}
	}
	unbinders = append(unbinders, func() error {
		return sub.Unsubscribe()
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
		for r.buffer.Len() < len(p) {
			bodyChunkMsg, err := r.natsSubscription.NextMsg(DispatchTimeout)
			if err != nil {
				return 0, err
			}
			bodyChunk := &BodyChunk{}
			if err := msgpack.Unmarshal(bodyChunkMsg.Data, bodyChunk); err != nil {
				return 0, err
			}

			if bodyChunk.IsEOF {
				r.hasEnded = true
				if err := r.natsSubscription.Unsubscribe(); err != nil {
					return 0, err
				}
				break
			}

			if _, err := r.buffer.Write(bodyChunk.Data); err != nil {
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
	n, err := r.buffer.Write(p)
	if err != nil {
		return 0, err
	}
	chunkCount := int(math.Floor(float64(r.buffer.Len()) / DispatchBodyChunkSize))
	for i := 0; i < chunkCount; i += 1 {
		if err := r.writeChunk(); err != nil {
			return 0, err
		}
	}
	return n, nil
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
	bodyChunkBytes, err := msgpack.Marshal(bodyChunk)
	if err != nil {
		return err
	}

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
	responseBytes, err := msgpack.Marshal(response)
	if err != nil {
		return err
	}

	return r.natsConnection.Publish(r.responseSubject, responseBytes)
}

func (r *responseWriter) writeChunk() error {
	bodyChunkBytes, err := msgpack.Marshal(&BodyChunk{
		Index: r.writeIndex,
		Data:  r.buffer.Next(DispatchBodyChunkSize),
	})
	if err != nil {
		return err
	}
	if err := r.natsConnection.Publish(r.responseBodySubject, bodyChunkBytes); err != nil {
		return err
	}
	r.writeIndex += 1
	return nil
}

func (c *NatsTransport) handleDispatch(msg *nats.Msg, handler func(res http.ResponseWriter, req *http.Request)) error {
	responseBodySubject := nats.NewInbox()

	request := &Request{}
	if err := msgpack.Unmarshal(msg.Data, request); err != nil {
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
	requestAckBytes, err := msgpack.Marshal(requestAck)
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

func (c *NatsTransport) UnbindDispatch(serviceName string) error {
	unbinders, ok := c.unbindDispatch[serviceName]
	if ok {
		for _, unbind := range unbinders {
			if err := unbind(); err != nil {
				return err
			}
		}
	}
	return nil
}

type eofReader struct{}

var _ io.ReadCloser = &eofReader{}

func (r *eofReader) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (r *eofReader) Close() error {
	return nil
}
