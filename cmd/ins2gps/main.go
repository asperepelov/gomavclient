package main

import (
	"flag"
	"fmt"
	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
	"gomavclient/mavlink"
	"log"
	"sync"
	"time"
)

func main() {
	var err error

	// Объявление параметров со значениями по умолчанию и описанием
	protocol := flag.String("protocol", "tcp", "Протокол телеметрии для использования (tcp или udp)")
	address := flag.String("address", "", "IP адрес и порт телеметрии (например, 192.168.144.100:5600)")
	serialPort := flag.String("serial", "", "Последовательный порт для передачи GPS и INS (например, com12 или /dev/ttyAMA1)")
	serialBaud := flag.Int("serial_baud", 115200, "Скорость передачи серийного порта (baud rate) (например, 115200)")

	// Разбор аргументов
	flag.Parse()

	// Валидация protocol
	if *protocol != "tcp" && *protocol != "udp" {
		log.Fatalf("Не известный протокол: %s, должно быть tcp или udp", *protocol)
	}

	// Проверка обязательных параметров
	if *address == "" && *serialPort == "" {
		log.Fatalf("Параметры -address и -serial должны быть заданы")
	}

	fmt.Println("Используемые параметры:")
	fmt.Println("Протокол телеметрии:", *protocol)
	fmt.Println("Адрес телеметрии:", *address)
	fmt.Println("Последовательный порт для передачи:", *serialPort)
	fmt.Println("Скорость передачи последовательного порта:", *serialBaud)

	telemManager := mavlink.NewTelemetryManager()

	// Настройка клиентского соединения mavlink
	var clientEndpointConf gomavlib.EndpointConf
	if *protocol == "tcp" {
		clientEndpointConf = gomavlib.EndpointTCPClient{Address: *address}
	} else if *protocol == "udp" {
		clientEndpointConf = gomavlib.EndpointUDPClient{Address: *address}
	}
	clientConn := mavlink.NewConnection(
		clientEndpointConf,
		10,
		mavlink.WithTelemetryManager(telemManager), // Обработка mavlink
		mavlink.WithDebug(true),
	)
	err = clientConn.Open()
	if err != nil {
		panic(err)
	}
	defer clientConn.Close()

	// Serial соединение
	serialEndpointConf := gomavlib.EndpointSerial{Device: *serialPort, Baud: *serialBaud} // Serial конфигурация
	serialNode := gomavlib.Node{
		Endpoints:        []gomavlib.EndpointConf{serialEndpointConf},
		Dialect:          ardupilotmega.Dialect,
		OutVersion:       gomavlib.V2,
		OutSystemID:      10,
		HeartbeatDisable: true,
	}
	err = serialNode.Initialize()
	if err != nil {
		log.Fatalf("Ошибка открытия последовательного порта: %v", err)
	}
	defer serialNode.Close()

	// Обработчик пакета ИНС
	telemManager.RegisterCallback(&common.MessageGlobalPositionInt{}, func(msg interface{}) {
		ins := msg.(*common.MessageGlobalPositionInt)
		//fmt.Printf("Получен ИНС пакет: %+v\n", ins)

		// Пишем ИНС в serial
		err := serialNode.WriteMessageAll(ins)
		if err != nil {
			log.Printf("Ошибка отправки в Serial порт: %v\n", err)
		}

		// Формирование пакета Gps
		newGps := common.MessageGpsRawInt{
			TimeUsec:          mavlink.GetTimeUsec(),
			FixType:           3,
			Lat:               ins.Lat,
			Lon:               ins.Lon,
			Alt:               ins.Alt,
			Eph:               0,
			Epv:               0,
			Vel:               0,
			Cog:               0,
			SatellitesVisible: 32,
			AltEllipsoid:      0,
			HAcc:              16000,
			VAcc:              16000,
			VelAcc:            16000,
			HdgAcc:            24000,
			Yaw:               0,
		}

		// Пишем GPS в serial
		err = serialNode.WriteMessageAll(&newGps)
		if err != nil {
			log.Printf("Ошибка отправки в Serial порт: %v\n", err)
		}
		log.Printf("Отправлен Gps %+v\n", newGps)
	})

	//////////////////////////////////////////
	// Запуск горутин
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Горутина для обработки mavlink сообщений
	go func() {
		defer wg.Done()
		clientConn.HandleEvents()
	}()

	// Статус сервиса
	go func() {
		for {
			defer wg.Done()
			fmt.Println(clientConn.Info())
			time.Sleep(15 * time.Minute)
		}
	}()

	wg.Wait()
}
