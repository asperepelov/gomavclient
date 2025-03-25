package actions

import (
	"fmt"
	"gomavclient/mavlink"
	"log"
)

type ActionManager struct {
	connection *mavlink.Connection

	goGoStartLat float32
	goGoStartLon float32
	goGoStartAlt float32
}

func NewActionManager(connection *mavlink.Connection) *ActionManager {
	return &ActionManager{
		connection: connection,
	}
}

func (m *ActionManager) GoGo(goParamId string, goEnable float32) {
	if m.connection.TelemetryManager == nil {
		log.Printf("Ошибка! Отсутствует TelemetryManager")
		return
	}

	m.goGoStartLat = float32(m.connection.TelemetryManager.GlobalPositionInt.Lat)
	m.goGoStartLon = float32(m.connection.TelemetryManager.GlobalPositionInt.Lon)
	m.goGoStartAlt = float32(m.connection.TelemetryManager.GlobalPositionInt.Alt)

	if goEnable == 1 {
		fmt.Println("GoGo started")
	} else if goEnable == 2 {
		fmt.Println("GoGo new waypoint")
		alt := float32(0)
		if m.goGoStartAlt-50 > 0 {
			alt = m.goGoStartAlt - 50
		}
		err := m.connection.Write(mavlink.GetMissionItem(&mavlink.GeoPoint{Lat: 50.0, Lng: 10.0, Alt: alt}))
		if err != nil {
			fmt.Printf("GoGo write MissionItem error : %v", err)
		}
		// Передача управления скрипту
		err = m.connection.Write(mavlink.GetMessageParamSet(goParamId, 3))
		if err != nil {
			fmt.Printf("GoGo param set: %v", err)
		}
	} else if goEnable == 4 {
		fmt.Println("GoGo restore waypoint")
		err := m.connection.Write(mavlink.GetMissionItem(&mavlink.GeoPoint{Lat: 50.0, Lng: 10.0, Alt: m.goGoStartAlt}))
		if err != nil {
			log.Printf("GoGo write MissionItem error : %v", err)
		}
	}
}
