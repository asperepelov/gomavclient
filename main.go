package main

import (
	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"gomavclient/mavlink"
	"sync"
)

// this example shows how to:
// 1) create a node which communicates with a serial endpoint
// 2) print selected incoming messages

func main() {
	//endpointConf := gomavlib.EndpointSerial{Device: "com4", Baud: 57600} // Serial конфигурация
	//endpointConf := gomavlib.EndpointTCPClient{"127.0.0.1:5601"} // TCP конфигурация
	//connection := mavlink.NewConnection(endpointConf)
	//err := connection.Open()
	//if err != nil {
	//	panic(err)
	//}
	//defer connection.Close()

	node, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPClient{"127.0.0.1:5601"},
			//gomavlib.EndpointSerial{Device: "com4", Baud: 57600},
		},
		Dialect:     ardupilotmega.Dialect,
		OutVersion:  gomavlib.V2, // change to V1 if you're unable to communicate with the target
		OutSystemID: 10,
	})
	if err != nil {
		panic(err)
	}
	defer node.Close()

	wg := sync.WaitGroup{}
	wg.Add(2)

	// Горутина для обработки входящих сообщений
	go func() {
		defer wg.Done()
		//connection.HandleEvents()
		mavlink.HandleEvents(node)
	}()

	wg.Wait()

}
