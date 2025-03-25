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
	// Параметры
	params := mavlink.NewParamManager()
	ScrRebEnable := params.Register("SCR_REB_ENBL")
	ScrRebEnable.AddCallback(func(value float32) {
		fmt.Println("callback SCR_REB_ENBL:", value)
	})
	ScrRebMode := params.Register("SCR_REB_MODE")
	ScrRebMode.AddCallback(func(value float32) {
		fmt.Println("callback SCR_REB_MODE:", value)
	})
	goGoEnableParamId := "SCR_GOGO_ENBL"
	ScrGoGoEnable := params.Register(goGoEnableParamId)

	// Настройка соединения
	//endpointConf := gomavlib.EndpointSerial{Device: "com4", Baud: 57600} // Serial конфигурация
	endpointConf := gomavlib.EndpointTCPClient{"127.0.0.1:5601"} // TCP конфигурация
	connection := mavlink.NewConnection(
		endpointConf,
		10,
		mavlink.WithParamManager(params),
		mavlink.WithDebug(true),
	)
	err := connection.Open()
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	// Действия при обновлении параметра
	actionsManager := actions.NewActionManager(connection)
	ScrGoGoEnable.AddCallback(func(value float32) {
		fmt.Printf("GoGo %s: %.0f\n", goGoEnableParamId, value)
		actionsManager.GoGo(goGoEnableParamId, value)
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
