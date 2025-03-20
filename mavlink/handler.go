package mavlink

import (
	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"log"
	"time"
)

// HandleEvents обработка событий
func (c *Connection) HandleEvents() {
	for event := range c.node.Events() {
		switch ee := event.(type) {

		case *gomavlib.EventParseError: // Ошибка данных
			c.handleParseError(*ee)
			continue

		case *gomavlib.EventChannelOpen: // Открытие канала
			c.handleChannelOpen(*ee)

		case *gomavlib.EventChannelClose: // Закрытие канала
			c.handleChannelClose(*ee)
			continue
		}

		// Событие кадр данных
		if frm, ok := event.(*gomavlib.EventFrame); ok {
			switch msg := frm.Message().(type) {

			case *ardupilotmega.MessageHeartbeat:
				c.handleHeartbeat(msg)

			//case *ardupilotmega.MessageServoOutputRaw:
			//log.Printf("received servo output with values: %d %d %d %d %d %d %d %d\n",
			//	msg.Servo1Raw, msg.Servo2Raw, msg.Servo3Raw, msg.Servo4Raw,
			//	msg.Servo5Raw, msg.Servo6Raw, msg.Servo7Raw, msg.Servo8Raw)

			case *ardupilotmega.MessageParamValue:
				c.handleParamValue(msg)
			}
		}
	}
}

func (c *Connection) handleChannelClose(event gomavlib.EventChannelClose) {
	c.opened = false
	log.Printf("channel closed: %v\n", event)
}

func (c *Connection) handleChannelOpen(event gomavlib.EventChannelOpen) {
	c.opened = true
	log.Printf("channel opened: %v\n", event)
}

func (c *Connection) handleParseError(event gomavlib.EventParseError) {
	c.parseErrorCounter++
	log.Printf("parse error: %v\n", event)
}

func (c *Connection) handleHeartbeat(msg *ardupilotmega.MessageHeartbeat) {
	c.lastHeartbeat = time.Now()
}

func (c *Connection) handleParamValue(msg *ardupilotmega.MessageParamValue) {
	if c.paramManager != nil {
		c.paramManager.Update(msg.ParamId, msg.ParamValue)
	}
}
