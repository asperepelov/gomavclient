package main

import (
	"fmt"
	"github.com/bluenviron/gomavlib/v3"
	"gomavclient/mavlink"
	"sync"
	"time"
)

func main() {
	//endpointConf := gomavlib.EndpointSerial{Device: "com4", Baud: 57600} // Serial конфигурация
	endpointConf := gomavlib.EndpointTCPClient{"127.0.0.1:5601"} // TCP конфигурация
	connection := mavlink.NewConnection(endpointConf)
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
