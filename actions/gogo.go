package actions

import (
	"fmt"
	"gomavclient/mavlink"
	"log"
	"math"
)

type GoGo struct {
	connection *mavlink.Connection

	goGoStartLat      float32
	goGoStartLon      float32
	goGoStartAlt      float32
	goGoNeedRestoreWP bool
}

func NewGoGo(connection *mavlink.Connection) *GoGo {
	return &GoGo{
		connection: connection,
	}
}

func (m *GoGo) Run(goParamId string, goEnable float32) {
	if m.connection.TelemetryManager == nil {
		log.Printf("Ошибка! Отсутствует TelemetryManager")
		return
	}

	m.goGoStartLat = float32(m.connection.TelemetryManager.GlobalPositionInt.Lat)
	m.goGoStartLon = float32(m.connection.TelemetryManager.GlobalPositionInt.Lon)
	m.goGoStartAlt = float32(math.Round(float64(m.connection.TelemetryManager.VfrHud.Alt)))

	if goEnable == 1 {
		fmt.Println("GoGo started")
	} else if goEnable == 2 {
		fmt.Println("GoGo new waypoint")
		alt := float32(0)
		if m.goGoStartAlt-50 > 0 {
			alt = m.goGoStartAlt - 50
		}
		m.goGoNeedRestoreWP = true

		// начать манёвр
		msg := mavlink.GetMissionItem(&mavlink.GeoPoint{Lat: 10.100005, Lng: 10.000003, Alt: alt})
		fmt.Printf("GoGo new waypoint: %v", msg)
		err := m.connection.Write(msg)
		if err != nil {
			fmt.Printf("GoGo write MissionItem error : %v", err)
		}
		// Передача управления скрипту
		err = m.connection.Write(mavlink.GetMessageParamSet(goParamId, 3))
		if err != nil {
			fmt.Printf("GoGo param set: %v", err)
		}
	} else if m.goGoNeedRestoreWP && goEnable == 0 {
		fmt.Println("GoGo restore waypoint")
		m.goGoNeedRestoreWP = false

		// выход из манёвра
		msg := mavlink.GetMissionItem(&mavlink.GeoPoint{Lat: 10.000001, Lng: 10.000005, Alt: m.goGoStartAlt})
		fmt.Printf("GoGo restore waypoint: %v", msg)
		err := m.connection.Write(msg)
		if err != nil {
			log.Printf("GoGo write MissionItem error : %v", err)
		}
	}
}
