package main

import (
	"fmt"
	"github.com/bluenviron/gomavlib/v3"
	"gomavclient/actions"
	"gomavclient/mavlink"
	"sync"
	"time"
)

func main() {
	// Телеметрия
	telemetry := mavlink.NewTelemetryManager()

	// Параметры
	params := mavlink.NewParamManager()
	goGoEnableParamId := "SCR_GOGO_ENBL"
	ScrGoGoEnable := params.Register(goGoEnableParamId)

	// Настройка соединения
	//endpointConf := gomavlib.EndpointSerial{Device: "com4", Baud: 57600} // Serial конфигурация
	endpointConf := gomavlib.EndpointTCPClient{"127.0.0.1:5600"} // TCP конфигурация
	connection := mavlink.NewConnection(
		endpointConf,
		10,
		mavlink.WithParamManager(params),
		mavlink.WithTelemetryManager(telemetry),
		mavlink.WithDebug(true),
	)
	err := connection.Open()
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	// Действия при обновлении параметра
	goGo := actions.NewGoGo(connection)
	ScrGoGoEnable.AddCallback(func(value float32) {
		fmt.Printf("GoGo %s: %.0f\n", goGoEnableParamId, value)
		goGo.Run(goGoEnableParamId, value)
	})

	// Запуск горутин
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Горутина для обработки входящих сообщений
	go func() {
		defer wg.Done()
		connection.HandleEvents()
	}()

	go func() {
		for {
			defer wg.Done()
			fmt.Println(connection.Info())
			time.Sleep(5 * time.Second)
		}
	}()

	wg.Wait()
}
