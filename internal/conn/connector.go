package conn

import (
	"github.com/ferossa/mockston/internal/cfg"
	"log"
)

const (
	protocolHttp = "http"
	protocolAmqp = "amqp"
)

type MessageHandler func(string, []byte, map[string]interface{}) ([]byte, error)

// IConnector interface for connectors
type IConnector interface {
	SetEndpoints([]cfg.Endpoint, MessageHandler) error
	Connect() error
}

// NewConnector create new connector
func NewConnector(config cfg.Connection) IConnector {
	switch config.Protocol {
	case protocolHttp:
		return &HttpConnector{config: config}
	case protocolAmqp:
		return &AmqpConnector{config: config}
	}

	log.Fatalln("unknown connection protocol", config.Protocol)
	return nil
}
