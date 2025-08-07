package main

import (
	"fmt"
	"github.com/bluenviron/gomavlib/v3"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gomavclient/actions"
	"gomavclient/common"
	"gomavclient/mavlink"
	"log"
	"sync"
	"time"
)

const (
	broker     = "tcp://localhost:1883"
	clientID   = "svc-shield"
	paramPub   = "svc/shield/param/out"  // Publication
	commandSub = "svc/shield/command/in" // Subscription
)

func main() {
	// MQTT
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(clientID)
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Параметры mavlink
	mavParams := common.NewParamManager()

	// Настройка соединения mavlink
	//endpointConf := gomavlib.EndpointSerial{Device: "com4", Baud: 57600} // Serial конфигурация
	endpointConf := gomavlib.EndpointTCPClient{"127.0.0.1:5600"} // TCP конфигурация
	connection := mavlink.NewConnection(
		endpointConf,
		10,
		mavlink.WithParamManager(mavParams),
		mavlink.WithTelemetryManager(mavlink.NewTelemetryManager()), // Обработка mavlink
		mavlink.WithDebug(true),
	)
	err := connection.Open()
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	// Отслеживаемые параметры mavlink
	parScanEnable := mavParams.Register("ANDR_SCAN_ENBL", common.WithUploadStartup(true))
	parRebAuto := mavParams.Register("ANDR_REB_AUTO", common.WithUploadStartup(true))
	parGoGoAuto := mavParams.Register("ANDR_GOGO_AUTO", common.WithUploadStartup(true))
	parRebRun := mavParams.Register("ANDR_REB_RUN", common.WithUploadStartup(true))
	parAndrGoGoRun := mavParams.Register("ANDR_GOGO_RUN", common.WithUploadStartup(true))
	parAndrCourseDeg := mavParams.Register("ANDR_COURSEDEG", common.WithUploadStartup(true))
	parAndrChangeAlt := mavParams.Register("ANDR_CHANGEALT", common.WithUploadStartup(true))

	// Публикация значения параметра
	sendParamMqtt := func(param *common.Param) {
		json, _ := param.JSON()
		mqttClient.Publish(paramPub, 1, false, json)
	}

	// Действия при обновлении параметра
	parRebAuto.AddCallback(func(float32) { sendParamMqtt(parRebAuto) })
	parGoGoAuto.AddCallback(func(float32) { sendParamMqtt(parGoGoAuto) })
	parRebRun.AddCallback(func(float32) { sendParamMqtt(parRebRun) })
	var changeAlt float32
	parAndrChangeAlt.AddCallback(func(f float32) { changeAlt = f })
	var courseDeg float32
	parAndrCourseDeg.AddCallback(func(f float32) { courseDeg = f })

	// Вкл / Выкл RF
	rfSwitchOn := actions.NewRFSwitchOn("rf.service")
	parScanEnable.AddCallback(func(value float32) {
		rfSwitchOn.HandleParamValue(value)
		sendParamMqtt(parScanEnable)
	})

	// Вкл / Выкл GoGo
	goGo := actions.NewGoGo(connection, &changeAlt, &courseDeg)
	parAndrGoGoRun.AddCallback(func(value float32) {
		goGo.HandleParamValue(parAndrGoGoRun.Name, value)
		sendParamMqtt(parAndrGoGoRun)
	})

	// Обработка получения команд
	commandHandler := func(client mqtt.Client, msg mqtt.Message) {
		cmd := string(msg.Payload())
		log.Printf("Получена команда: %s\n", cmd)
		if cmd == "gogo" && parAndrGoGoRun.Value == 0 {
			goGo.Enable(parAndrGoGoRun.Name)
		}
	}
	mqttClient.Subscribe(commandSub, 1, commandHandler)

	//////////////////////////////////////////
	// Запуск горутин
	wg := sync.WaitGroup{}
	wg.Add(4)

	// Горутина для обработки mavlink сообщений
	go func() {
		defer wg.Done()
		connection.HandleEvents()
	}()

	// Горутина для стартовой загрузки параметров
	go func() {
		defer wg.Done()
		for {
			params := mavParams.GetParamsToUploadStartup()
			if len(params) == 0 {
				break
			}

			for _, param := range params {
				fmt.Println("Обновление при старте", param.Name)
				err := connection.Write(mavlink.GetMessageParamRequestRead(param.Name))
				if err != nil {
					fmt.Printf("%v\n", err)
				}
			}
			time.Sleep(3 * time.Second)
		}
	}()

	// Горутина для обновления параметров
	go func() {
		defer wg.Done()
		for {
			params := mavParams.GetParamsToRefresh()

			for _, param := range params {
				fmt.Println("Пора обновить", param.Name)
				err := connection.Write(mavlink.GetMessageParamRequestRead(param.Name))
				if err != nil {
					fmt.Printf("%v\n", err)
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Статус сервиса
	go func() {
		for {
			defer wg.Done()
			fmt.Println(connection.Info())
			time.Sleep(15 * time.Minute)
		}
	}()

	wg.Wait()
}
