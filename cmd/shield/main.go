package main

import (
	"fmt"
	"github.com/bluenviron/gomavlib/v3"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gomavclient/actions"
	"gomavclient/common"
	"gomavclient/mavlink"
	"sync"
	"time"
)

const (
	broker    = "tcp://localhost:1883"
	clientID  = "svc-shield"
	paramPub  = "svc/shield/param"     // Publication
	dangerSub = "svc/freq_scan/danger" // Subscription
)

func main() {
	// MQTT
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(clientID)
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	// Подписка на тревогу
	var dangerHandler mqtt.MessageHandler
	mqttClient.Subscribe(dangerSub, 1, dangerHandler)

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
	parScanEnable := mavParams.Register("ANDR_SCAN_ENBL")
	parRebAuto := mavParams.Register("ANDR_REB_AUTO")
	parGoGoAuto := mavParams.Register("ANDR_GOGO_AUTO")
	parRebRun := mavParams.Register("ANDR_REB_RUN")
	parAndrGoGoRun := mavParams.Register("ANDR_GOGO_RUN")

	// Публикация значения параметра
	sendParamMqtt := func(param *common.Param) {
		json, _ := param.JSON()
		mqttClient.Publish(paramPub, 1, false, json)
	}

	// Действия при обновлении параметра
	parScanEnable.AddCallback(func(float32) { sendParamMqtt(parScanEnable) })
	parRebAuto.AddCallback(func(float32) { sendParamMqtt(parRebAuto) })
	parGoGoAuto.AddCallback(func(float32) { sendParamMqtt(parGoGoAuto) })
	parRebRun.AddCallback(func(float32) { sendParamMqtt(parRebRun) })

	goGo := actions.NewGoGo(connection)
	parAndrGoGoRun.AddCallback(func(value float32) {
		fmt.Printf("GoGo %s: %.0f\n", parAndrGoGoRun.Name, value)
		goGo.HandleParamValue(parAndrGoGoRun.Name, value)
		sendParamMqtt(parAndrGoGoRun)
	})

	// Обработка получения тревоги
	dangerHandler = func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Получена тревога: %s\n", msg.Payload())
		goGo.Enable(parAndrGoGoRun.Name)
	}

	// Запуск горутин
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Горутина для обработки mavlink сообщений
	go func() {
		defer wg.Done()
		connection.HandleEvents()
	}()

	go func() {
		for {
			defer wg.Done()
			fmt.Println(connection.Info())
			time.Sleep(1 * time.Hour)
		}
	}()

	wg.Wait()
}
