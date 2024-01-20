package nats

import (
	"strings"
	"time"
	"unicode"

	"github.com/nats-io/nats.go"
	"github.com/telemetrytv/zephyr"
)

const DispatchTimeout = 30 * time.Second
const NatsSubjectNamespace = "zephyr"

func NewConnection(natsConn *nats.Conn) zephyr.Connection {
	return &Connection{
		Connection: natsConn,
	}
}

type Connection struct {
	Connection                           *nats.Conn
	onUnbindAnnounceGateway              func()
	onUnbindAnnounceService              func()
	onUnBindDispatchFromGatewayOrService func()
}

type RequestMessage struct {
	Method                    string              `json:"method"`
	URL                       string              `json:"url"`
	Params                    map[string]string   `json:"params"`
	Headers                   map[string][]string `json:"headers"`
	ResponseSubject           string              `json:"responseSubject"`
	ResponseBodyStreamSubject string              `json:"responseBodyStreamSubject"`
}

type RequestAckMessage struct {
	BodyStreamSubject string `json:"bodyStreamSubject"`
}

type ResponseMessage struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
}

type BodyChunkMessage struct {
	Index int    `json:"index"`
	Data  []byte `json:"data"`
	EOF   bool   `json:"eof"`
}

func namespace(strValues ...string) string {
	namespaceChunks := []string{NatsSubjectNamespace}
	for _, str := range strValues {
		namespaceChunks = append(namespaceChunks, formatForNamespace(str))
	}
	return strings.Join(namespaceChunks, ".")
}

func formatForNamespace(str string) string {
	formattedStr := []rune{}
	for i, r := range str {
		pi := i - 1
		var pr rune
		if pi >= 0 {
			pr = []rune(str)[pi]
		}
		switch {
		case r >= 'A' && r <= 'Z':
			if pr >= 'a' && pr <= 'z' {
				formattedStr = append(formattedStr, '-', unicode.ToLower(r))
			} else {
				formattedStr = append(formattedStr, r)
			}
		case r >= 'a' && r <= 'z':
			formattedStr = append(formattedStr, r)
		case r >= '0' && r <= '9':
			formattedStr = append(formattedStr, r)
		case r == '-':
			formattedStr = append(formattedStr, '-')
		case r == '_':
			formattedStr = append(formattedStr, '-')
		case r == '.':
			formattedStr = append(formattedStr, '.')
		case r == '*':
			formattedStr = append(formattedStr, '*')
		}
	}
	return string(formattedStr)
}
