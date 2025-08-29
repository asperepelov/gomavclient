package actions

import (
	"fmt"
	"gomavclient/mavlink"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	earthRadius   = 6371.0 // радиус Земли в километрах
	distanceSide  = 0.5    // расстояние вбок в километрах
	distanceAhead = 1      // расстояние вперед в километрах
)

type GoGo struct {
	connection *mavlink.Connection
	mu         sync.Mutex
	step       uint8

	goGoStartHeading  int16
	goGoStartLat      float32
	goGoStartLon      float32
	goGoStartAlt      float32
	goGoNeedRestoreWP bool
	changeAlt         *float32
	courseDeg         *float32
}

func NewGoGo(connection *mavlink.Connection, _changeAlt *float32, _courseDeg *float32) *GoGo {
	return &GoGo{
		connection: connection,
		step:       0,
		changeAlt:  _changeAlt,
		courseDeg:  _courseDeg,
	}
}

// Enable активировать GoGo
func (m *GoGo) Enable(goParamId string) {
	// Значение параметра в 1
	err := m.connection.Write(mavlink.GetMessageParamSet(goParamId, 1))
	if err != nil {
		fmt.Printf("Error GoGo param set: %v", err)
	}
}

// HandleParamValue Обработка изменений параметра
func (m *GoGo) HandleParamValue(goParamId string, goEnable float32) {
	if m.connection.TelemetryManager == nil {
		log.Printf("Ошибка! Отсутствует TelemetryManager")
		return
	}

	if goEnable == 2 {
		m.mu.Lock()
		if m.step != 0 {
			m.mu.Unlock()
			return
		}
		m.step = 2

		fmt.Println("GoGo started")
		m.goGoStartHeading = m.connection.TelemetryManager.Heading
		m.goGoStartLat = m.connection.TelemetryManager.Lat
		m.goGoStartLon = m.connection.TelemetryManager.Lon
		m.goGoStartAlt = float32(math.Round(float64(m.connection.TelemetryManager.VfrHud.Alt)))

		fmt.Println("GoGo new waypoint")
		alt := float32(0)
		if m.goGoStartAlt-*m.changeAlt > 0 {
			alt = m.goGoStartAlt - *m.changeAlt
		}
		m.goGoNeedRestoreWP = true

		// точка в стороне
		sideLat, sideLon, side := CalculatePointSide(
			float64(m.goGoStartHeading),
			float64(m.goGoStartLat),
			float64(m.goGoStartLon),
			float64(*m.courseDeg),
			distanceSide)
		log.Printf("Манёвр %s со снижением на %d и отклонением от курса на %d град", side, int(*m.changeAlt), int(*m.courseDeg))

		msg := mavlink.GetMissionItem(&mavlink.GeoPoint{Lat: sideLat, Lng: sideLon, Alt: alt})
		fmt.Printf("GoGo new waypoint: %v\n", msg)
		err := m.connection.Write(msg)
		if err != nil {
			fmt.Printf("GoGo write MissionItem error : %v\n", err)
		}
		// Передача управления скрипту
		err = m.connection.Write(mavlink.GetMessageParamSet(goParamId, 3))
		if err != nil {
			fmt.Printf("Error GoGo param set: %v\n", err)
		}

		m.mu.Unlock()
	} else if m.goGoNeedRestoreWP && goEnable == 0 {
		m.mu.Lock()
		if m.step != 2 {
			m.mu.Unlock()
			return
		}
		m.step = 0

		m.goGoNeedRestoreWP = false

		// точка впереди по курсу
		aheadLat, aheadLon := CalculatePointAhead(
			float64(m.goGoStartHeading),
			float64(m.goGoStartLat),
			float64(m.goGoStartLon),
			distanceAhead)

		// выход из манёвра
		msg := mavlink.GetMissionItem(&mavlink.GeoPoint{Lat: aheadLat, Lng: aheadLon, Alt: m.goGoStartAlt})
		fmt.Printf("GoGo restore waypoint: %v", msg)
		err := m.connection.Write(msg)
		if err != nil {
			log.Printf("GoGo write MissionItem error : %v", err)
		}

		m.mu.Unlock()
	}
}

// CalculatePointAhead вычисляет координаты точки, находящейся на заданном расстоянии
// впереди по курсу от текущей позиции
func CalculatePointAhead(heading, lat, lon, distance float64) (float32, float32) {
	// Перевод градусов в радианы
	lat1 := lat * math.Pi / 180.0
	lon1 := lon * math.Pi / 180.0
	headingRad := heading * math.Pi / 180.0

	// Расчет расстояния в радианах
	angularDistance := distance / earthRadius

	// Расчет новой широты
	lat2 := math.Asin(math.Sin(lat1)*math.Cos(angularDistance) +
		math.Cos(lat1)*math.Sin(angularDistance)*math.Cos(headingRad))

	// Расчет новой долготы
	lon2 := lon1 + math.Atan2(math.Sin(headingRad)*math.Sin(angularDistance)*math.Cos(lat1),
		math.Cos(angularDistance)-math.Sin(lat1)*math.Sin(lat2))

	// Перевод обратно в градусы
	lat2Deg := lat2 * 180.0 / math.Pi
	lon2Deg := lon2 * 180.0 / math.Pi

	return float32(lat2Deg), float32(lon2Deg)
}

// CalculatePointSide вычисляет координаты точки, находящейся на заданном расстоянии
// слева или справа от курса с отклонением на courseDeg
func CalculatePointSide(heading, lat, lon, distance float64, courseDeg float64) (float32, float32, string) {
	// Инициализация генератора случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Случайный выбор стороны (0 - слева, 1 - справа)
	side := rand.Intn(2)
	sideDescription := "вправо"

	// Если левая сторона, корректируем курс на courseDeg градусов против часовой стрелки
	// Если правая сторона, корректируем курс на courseDeg градусов по часовой стрелке
	if side == 0 {
		heading = math.Mod(heading-courseDeg+360, 360)
		sideDescription = "влево"
	} else {
		heading = math.Mod(heading+courseDeg, 360)
	}

	// Используем ту же функцию для расчета точки по новому направлению
	lat2, lon2 := CalculatePointAhead(heading, lat, lon, distance)

	return lat2, lon2, sideDescription
}
