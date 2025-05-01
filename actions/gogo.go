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
	earthRadius = 6371.0 // радиус Земли в километрах
	distance    = 30.0   // расстояние в километрах
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
}

func NewGoGo(connection *mavlink.Connection) *GoGo {
	return &GoGo{
		connection: connection,
		step:       0,
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

	if goEnable == 1 {
		m.mu.Lock()
		if m.step != 0 {
			m.mu.Unlock()
			return
		}
		m.step = 1

		fmt.Println("GoGo started")
		m.goGoStartHeading = m.connection.TelemetryManager.Heading
		m.goGoStartLat = m.connection.TelemetryManager.Lat
		m.goGoStartLon = m.connection.TelemetryManager.Lon
		m.goGoStartAlt = float32(math.Round(float64(m.connection.TelemetryManager.VfrHud.Alt)))

		m.mu.Unlock()
	} else if goEnable == 2 {
		m.mu.Lock()
		if m.step != 1 {
			m.mu.Unlock()
			return
		}
		m.step = 2

		fmt.Println("GoGo new waypoint")
		alt := float32(0)
		if m.goGoStartAlt-50 > 0 {
			alt = m.goGoStartAlt - 50
		}
		m.goGoNeedRestoreWP = true

		// точка в стороне
		sideLat, sideLon, side := CalculatePointSide(
			float64(m.goGoStartHeading),
			float64(m.goGoStartLat),
			float64(m.goGoStartLon))
		log.Println("Манёвр", side)

		msg := mavlink.GetMissionItem(&mavlink.GeoPoint{Lat: sideLat, Lng: sideLon, Alt: alt})
		fmt.Printf("GoGo new waypoint: %v", msg)
		err := m.connection.Write(msg)
		if err != nil {
			fmt.Printf("GoGo write MissionItem error : %v", err)
		}
		// Передача управления скрипту
		err = m.connection.Write(mavlink.GetMessageParamSet(goParamId, 3))
		if err != nil {
			fmt.Printf("Error GoGo param set: %v", err)
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
			float64(m.goGoStartLon))

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
func CalculatePointAhead(heading, lat, lon float64) (float32, float32) {
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
// слева или справа от курса
func CalculatePointSide(heading, lat, lon float64) (float32, float32, string) {
	// Инициализация генератора случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Случайный выбор стороны (0 - слева, 1 - справа)
	side := rand.Intn(2)
	sideDescription := "вправо"

	// Если левая сторона, корректируем курс на 90 градусов против часовой стрелки
	// Если правая сторона, корректируем курс на 90 градусов по часовой стрелке
	if side == 0 {
		heading = math.Mod(heading-90+360, 360)
		sideDescription = "влево"
	} else {
		heading = math.Mod(heading+90, 360)
	}

	// Используем ту же функцию для расчета точки по новому направлению
	lat2, lon2 := CalculatePointAhead(heading, lat, lon)

	return lat2, lon2, sideDescription
}
