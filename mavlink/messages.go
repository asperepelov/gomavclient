package mavlink

import (
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
)

type GeoPoint struct {
	Lat float32 // Latitude целое число
	Lng float32 // Longitude целое число
	Alt float32 // Altitude в метрах
}

func GetMissionItem(wp *GeoPoint) *ardupilotmega.MessageMissionItem {
	return &ardupilotmega.MessageMissionItem{
		TargetSystem:    1,
		TargetComponent: 1,
		Seq:             0,
		Frame:           ardupilotmega.MAV_FRAME_GLOBAL_RELATIVE_ALT,
		Command:         common.MAV_CMD(ardupilotmega.MAV_CMD_NAV_WAYPOINT),
		Current:         2,
		Autocontinue:    1,
		Param1:          0,
		Param2:          0,
		Param3:          0,
		Param4:          0,
		X:               wp.Lat, // Latitude
		Y:               wp.Lng, // Longitude
		Z:               wp.Alt, // Altitude
		MissionType:     ardupilotmega.MAV_MISSION_TYPE_MISSION,
	}
}

func GetMessageParamSet(paramId string, value float32) *common.MessageParamSet {
	return &common.MessageParamSet{
		TargetSystem:    1,
		TargetComponent: 1,
		ParamId:         paramId,
		ParamValue:      value,
	}
}
