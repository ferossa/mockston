package app

import (
	"github.com/ferossa/mockston/internal/cfg"
	"github.com/ferossa/mockston/internal/conn"
)

// IListener listener interface
type IListener interface {
	Listen(listen cfg.Listen) error
}

// NewListener creates new listener
func NewListener(c conn.IConnector, p IProcessor) *Listener {
	return &Listener{
		c,
		p,
	}
}

// Listener listening for external requests
type Listener struct {
	Connection conn.IConnector
	Processor  IProcessor
}

// Listen start listening for data
func (l *Listener) Listen(conf cfg.Listen) error {
	if err := l.Connection.SetEndpoints(conf.Endpoints, l.onMessage); err != nil {
		return err
	}

	if err := l.Connection.Connect(); err != nil {
		return err
	}

	return nil
}

// onMessage process data received from connector
func (l *Listener) onMessage(endpoint string, data []byte, context map[string]interface{}) ([]byte, error) {
	rq := &ProcessRequest{
		Endpoint: endpoint,
		Content:  data,
		Context:  context,
	}

	resp, err := l.Processor.Process(rq)
	if err != nil {
		return nil, err
	}

	return resp.Content, nil
}
