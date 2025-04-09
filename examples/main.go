package main

import (
	"fmt"
	"github.com/bluenviron/gomavlib/v3"
	"gomavclient/common"
	"gomavclient/mavlink"
	"sync"
	"time"
)

func main() {
	// Параметры
	params := common.NewParamManager()
	parUser1 := params.Register("SCR_USER1")
	parUser1.AddCallback(func(value float32) {
		fmt.Println("callback SCR_USER1:", value)
	})
	parUser2 := params.Register("callback SCR_USER2")
	parUser2.AddCallback(func(value float32) {
		fmt.Println("SCR_USER2:", value)
	})

	// Настройка соединения
	//endpointConf := gomavlib.EndpointSerial{Device: "com4", Baud: 57600} // Serial конфигурация
	endpointConf := gomavlib.EndpointTCPClient{"127.0.0.1:5601"} // TCP конфигурация
	connection := mavlink.NewConnection(
		endpointConf,
		mavlink.WithParamManager(params),
		mavlink.WithDebug(true),
	)
	err := connection.Open()
	if err != nil {
		panic(err)
	}
	defer connection.Close()

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
