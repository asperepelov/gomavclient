package mavlink

import (
	"fmt"
	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"github.com/bluenviron/gomavlib/v3/pkg/message"
	"time"
)

type Connection struct {
	OutSystemID       byte
	opened            bool      // Соединение открыто
	parseErrorCounter int       // Счетчик ошибок парсинга
	lastHeartbeat     time.Time // Время последнего heartbeat

	endpointConf gomavlib.EndpointConf // Конфигурация соединения
	node         *gomavlib.Node

	// Опции
	paramManager     *ParamManager
	TelemetryManager *TelemetryManager
	debug            bool
}

func NewConnection(
	endpoint gomavlib.EndpointConf,
	outSystemID byte,
	options ...ConnectionOption,
) *Connection {
	conn := &Connection{
		endpointConf: endpoint,
		OutSystemID:  outSystemID,
	}

	// Применяем каждую опцию
	for _, option := range options {
		option(conn)
	}

	return conn
}

func (c *Connection) IsOpened() bool {
	return c.opened
}

func (c *Connection) LastHeartbeat() time.Time {
	return c.lastHeartbeat
}

func (c *Connection) ParseErrorCounter() int {
	return c.parseErrorCounter
}

func (c *Connection) Open() error {
	if c.node == nil {
		c.node = &gomavlib.Node{
			Endpoints:        []gomavlib.EndpointConf{c.endpointConf},
			Dialect:          ardupilotmega.Dialect,
			OutVersion:       gomavlib.V2,
			OutSystemID:      c.OutSystemID,
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

func (c *Connection) Info() string {
	return fmt.Sprintf("opened: %v, parse errors: %d, last heartbeat: %s", c.opened, c.parseErrorCounter, c.lastHeartbeat.Format(time.DateTime))
}
