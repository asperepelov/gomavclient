package mavlink

import (
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
)

type TelemetryManager struct {
	TimeBootMs        uint32
	Heading           int16                            // Курс в градусах
	GlobalPositionInt *common.MessageGlobalPositionInt // Положение ИНС
	VfrHud            *common.MessageVfrHud
}

func NewTelemetryManager() *TelemetryManager {
	return &TelemetryManager{}
}

func (t *TelemetryManager) HandleMessageGlobalPositionInt(msg *common.MessageGlobalPositionInt) {
	t.GlobalPositionInt = msg
	t.TimeBootMs = msg.TimeBootMs
	//fmt.Printf("Lat: %d, Lon: %d\n", t.GlobalPositionInt.Lat, t.GlobalPositionInt.Lon)
}

func (t *TelemetryManager) HandleMessageVfrHud(msg *common.MessageVfrHud) {
	t.VfrHud = msg
	t.Heading = t.VfrHud.Heading
	//fmt.Printf("Alt: %.2f, Heading: %d\n", t.VfrHud.Alt, t.VfrHud.Heading)
}
