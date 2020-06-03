package conn

import (
	"strconv"

	"github.com/streadway/amqp"

	"github.com/ferossa/mockston/internal/cfg"
)

// AmqpConnector connector to receive or send amqp messages
type AmqpConnector struct {
	config           cfg.Connection
	endpoints        []cfg.Endpoint
	connectionString string
	conn             *amqp.Connection
	handler          MessageHandler
}

// rebuildEndpoints rebuild amqp topology
func (c *AmqpConnector) rebuildEndpoints() error {
	for _, endpoint := range c.endpoints {
		ch, err := c.conn.Channel()
		if err != nil {
			return err
		}

		_, err = ch.QueueDeclare(
			endpoint.Queue,
			false,
			true,
			false,
			false,
			nil,
		)
		if err != nil {
			return err
		}

		// todo: exchange binding

		messages, err := ch.Consume(
			endpoint.Queue,
			"mockston",
			false,
			false,
			false,
			false,
			nil,
		)

		go func(ep *cfg.Endpoint) {
			for m := range messages {
				// process message
				resp, err := c.handler(ep.Name, m.Body, map[string]interface{}{})
				if err != nil {
					_ = m.Nack(false, false)
					continue
				}

				if m.ReplyTo != "" {
					_ = ch.Publish(
						"",
						m.ReplyTo,
						false,
						false,
						amqp.Publishing{
							ContentType:   "text/plain",
							CorrelationId: m.CorrelationId,
							Body:          resp,
						},
					)
				}

				_ = m.Ack(false)
			}
		}(&endpoint)
	}

	return nil
}

// SetEndpoints set endpoints to process
func (c *AmqpConnector) SetEndpoints(endpoints []cfg.Endpoint, h MessageHandler) error {
	c.handler = h
	c.endpoints = endpoints

	return nil
}

// Connect start listening for data
func (c *AmqpConnector) Connect() error {
	c.connectionString = "amqp://" + c.config.Login + ":" + c.config.Password + "@" + c.config.Host + ":" + strconv.FormatInt(int64(c.config.Port), 10)

	var err error
	c.conn, err = amqp.DialConfig(c.connectionString, amqp.Config{})
	if err != nil {
		return err
	}

	return c.rebuildEndpoints()
}
