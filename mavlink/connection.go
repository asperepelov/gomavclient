package mavlink

import (
	"fmt"
	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"github.com/bluenviron/gomavlib/v3/pkg/message"
)

type Connection struct {
	opened bool // Соединение открыто

	endpointConf gomavlib.EndpointConf // Конфигурация соединения
	node         *gomavlib.Node
}

func NewConnection(endpoint gomavlib.EndpointConf) *Connection {
	conn := &Connection{
		endpointConf: endpoint,
	}
	return conn
}

func (c *Connection) IsOpened() bool {
	return c.opened
}

func (c *Connection) Open() error {
	if c.node == nil {
		c.node = &gomavlib.Node{
			Endpoints:        []gomavlib.EndpointConf{c.endpointConf},
			Dialect:          ardupilotmega.Dialect,
			OutVersion:       gomavlib.V2,
			OutSystemID:      10,
			HeartbeatDisable: true,
		}
		err := c.node.Initialize()
		if err != nil {
			c.opened = false
			return fmt.Errorf("error creating mavlink node: %v", err)
		}
		c.opened = true
	}

	return nil
}

func (c *Connection) Close() {
	if c.node != nil {
		c.node.Close()
		c.node = nil
	}
	c.opened = false
}

func (c *Connection) Write(msg message.Message) error {
	if c.node != nil {
		err := c.node.WriteMessageAll(msg)
		if err != nil {
			return fmt.Errorf("Ошибка при отправке mavlink запроса: %s \n%v", msg, err)
		}
	}
	return nil
}
