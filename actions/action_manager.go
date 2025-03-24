package actions

import (
	"fmt"
	"gomavclient/mavlink"
	"log"
)

type ActionManager struct {
	connection *mavlink.Connection
}

func NewActionManager(connection *mavlink.Connection) *ActionManager {
	return &ActionManager{
		connection: connection,
	}
}

func (m *ActionManager) GoGo(goParamId string, goEnable float32) {
	if goEnable == 1 {
		fmt.Println("GoGo started")
	} else if goEnable == 2 {
		fmt.Println("GoGo new waypoint")
		err := m.connection.Write(mavlink.GetMissionItem(&mavlink.GeoPoint{Lat: 0, Lng: 0, Alt: 0}))
		if err != nil {
			log.Printf("GoGo write MissionItem error : %v", err)
		}
		// Передача управления скрипту
		err = m.connection.Write(mavlink.GetMessageParamSet(goParamId, 3))
		if err != nil {
			log.Printf("GoGo param set: %v", err)
		}
	} else if goEnable == 4 {
		fmt.Println("GoGo restore waypoint")
		err := m.connection.Write(mavlink.GetMissionItem(&mavlink.GeoPoint{Lat: 481229792, Lng: 376768832, Alt: 0}))
		if err != nil {
			log.Printf("GoGo write MissionItem error : %v", err)
		}
	}
}
