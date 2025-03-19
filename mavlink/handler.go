package mavlink

import (
	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"log"
)

// HandleEvents обработка событий
func (c *Connection) HandleEvents() {
	for event := range c.node.Events() {
		switch ee := event.(type) {

		case *gomavlib.EventParseError: // Ошибка данных
			handleParseError(*ee)
			continue

		case *gomavlib.EventChannelOpen: // Открытие канала
			handleChannelOpen(*ee)

		case *gomavlib.EventChannelClose: // Закрытие канала
			handleChannelClose(*ee)
			continue
		}

		// Событие кадр данных
		if frm, ok := event.(*gomavlib.EventFrame); ok {
			switch msg := frm.Message().(type) {

			case *ardupilotmega.MessageHeartbeat:
				handleHeartbeat(msg)

			// if frm.Message() is a *ardupilotmega.MessageServoOutputRaw, access its fields
			case *ardupilotmega.MessageServoOutputRaw:
				log.Printf("received servo output with values: %d %d %d %d %d %d %d %d\n",
					msg.Servo1Raw, msg.Servo2Raw, msg.Servo3Raw, msg.Servo4Raw,
					msg.Servo5Raw, msg.Servo6Raw, msg.Servo7Raw, msg.Servo8Raw)
			}
		}
	}
}

func HandleEvents(node *gomavlib.Node) {
	for event := range node.Events() {
		switch ee := event.(type) {

		case *gomavlib.EventParseError: // Ошибка данных
			handleParseError(*ee)
			continue

		case *gomavlib.EventChannelOpen: // Открытие канала
			handleChannelOpen(*ee)

		case *gomavlib.EventChannelClose: // Закрытие канала
			handleChannelClose(*ee)
			continue
		}

		// Событие кадр данных
		if frm, ok := event.(*gomavlib.EventFrame); ok {
			switch msg := frm.Message().(type) {

			case *ardupilotmega.MessageHeartbeat:
				handleHeartbeat(msg)

			// if frm.Message() is a *ardupilotmega.MessageServoOutputRaw, access its fields
			case *ardupilotmega.MessageServoOutputRaw:
				log.Printf("received servo output with values: %d %d %d %d %d %d %d %d\n",
					msg.Servo1Raw, msg.Servo2Raw, msg.Servo3Raw, msg.Servo4Raw,
					msg.Servo5Raw, msg.Servo6Raw, msg.Servo7Raw, msg.Servo8Raw)
			}
		}
	}
}

func handleChannelClose(event gomavlib.EventChannelClose) {
	log.Printf("channel closed: %v\n", event)
}

func handleChannelOpen(event gomavlib.EventChannelOpen) {
	log.Printf("channel opened: %v\n", event)
}

func handleParseError(event gomavlib.EventParseError) {
	log.Printf("parse error: %v\n", event)
}

func handleHeartbeat(event *ardupilotmega.MessageHeartbeat) {
	log.Printf("received heartbeat (type %d)\n", event.Type)
}
