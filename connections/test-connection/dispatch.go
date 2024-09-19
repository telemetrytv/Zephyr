package testconnection

import (
	"net/http"
	"time"
)

const Timeout = 30 * time.Second
const BodyChunkSize = 1024 * 16

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
	return nil
}

func (c *Connection) BindDispatch(serviceName string, handler func(res http.ResponseWriter, req *http.Request)) error {
	return nil
}

func (c *Connection) UnbindDispatch(serviceName string) {
}
