package mavlink

import (
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
	"reflect"
	"sync"
	"time"
)

type TelemetryManager struct {
	TimeBootMs        uint32
	Heading           int16                            // Курс в градусах
	Lat               float32                          // Широта в градусах
	Lon               float32                          // Долгота в градусах
	GlobalPositionInt *common.MessageGlobalPositionInt // Положение ИНС
	GpsRawInt         *common.MessageGpsRawInt         // Положение GPS
	VfrHud            *common.MessageVfrHud

	// map для хранения callbacks по типу сообщений
	callbacksMu sync.RWMutex
	callbacks   map[reflect.Type][]func(interface{})
}

func NewTelemetryManager() *TelemetryManager {
	return &TelemetryManager{
		callbacks: make(map[reflect.Type][]func(interface{})),
	}
}

// RegisterCallback Регистрация callback для типа сообщения
func (tm *TelemetryManager) RegisterCallback(msg interface{}, cb func(interface{})) {
	tm.callbacksMu.Lock()
	defer tm.callbacksMu.Unlock()
	t := reflect.TypeOf(msg)
	tm.callbacks[t] = append(tm.callbacks[t], cb)
}

// OnMessage Вызов callbacks при получении сообщения
func (tm *TelemetryManager) OnMessage(msg interface{}) {
	tm.callbacksMu.RLock()
	defer tm.callbacksMu.RUnlock()
	t := reflect.TypeOf(msg)
	for _, cb := range tm.callbacks[t] {
		go cb(msg) // вызвать callback асинхронно или синхронно по необходимости
	}
}

func (tm *TelemetryManager) HandleMessageGpsRawInt(msg *common.MessageGpsRawInt) {
	tm.GpsRawInt = msg
	//fmt.Printf("%+v\n", tm.GpsRawInt)
}

func (tm *TelemetryManager) HandleMessageGlobalPositionInt(msg *common.MessageGlobalPositionInt) {
	tm.GlobalPositionInt = msg
	tm.TimeBootMs = msg.TimeBootMs
	//fmt.Printf("Lat: %d, Lon: %d\n", tm.GlobalPositionInt.Lat, tm.GlobalPositionInt.Lon)
	tm.Lat = float32(msg.Lat) / 10000000.0
	tm.Lon = float32(msg.Lon) / 10000000.0
}

func (tm *TelemetryManager) HandleMessageVfrHud(msg *common.MessageVfrHud) {
	tm.VfrHud = msg
	tm.Heading = tm.VfrHud.Heading
	//fmt.Printf("Alt: %.2f, Heading: %d\n", tm.VfrHud.Alt, tm.VfrHud.Heading)
}

func GetTimeUsec() uint64 {
	now := time.Now()
	return uint64(now.UnixNano() / 1000)
}
