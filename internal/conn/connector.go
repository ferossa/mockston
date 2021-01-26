package conn

import (
	"log"

	"github.com/ferossa/mockston/internal/cfg"
)

const (
	protocolHttp = "http"
	protocolAmqp = "amqp"
)

// MessageHandler function called when message received
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
