package main

import (
	"fmt"
	"github.com/bluenviron/gomavlib/v3"
	"gomavclient/common"
	"gomavclient/mavlink"
	"gomavclient/raspi"
	"log"
	"sync"
	"time"
)

func main() {
	// Параметры
	params := common.NewParamManager()
	servoexp1 := params.Register("SERVOEXP1_PWM")
	servoexp2 := params.Register("SERVOEXP2_PWM")
	servoexp3 := params.Register("SERVOEXP3_PWM")
	servoexp4 := params.Register("SERVOEXP4_PWM")
	//servoexp5 := params.Register("SERVOEXP5_PWM")
	//servoexp6 := params.Register("SERVOEXP6_PWM")

	raspi.ListAvailablePins()

	// Инициализация аппаратного PWM
	pins := []raspi.Pin{
		//{Name: "GPIO12", InitValue: 1000},
		//{Name: "GPIO13", InitValue: 1000},
		{Name: "GPIO18", InitValue: 1000},
		//{Name: "GPIO19", InitValue: 1000},
	}
	hardwarePwm, err := raspi.NewHardwarePWMController(pins)
	if err != nil {
		log.Panic(fmt.Sprintf("Failed to create hardware PWM controller: %v", err))
	}

	// Добавление обработчиков значений параметров
	servoexp1.AddCallback(func(value float32) {
		err := hardwarePwm.SetPwm(raspi.ServoNumber(1), raspi.Pwm(value))
		if err != nil {
			log.Println(err)
		}
	})
	servoexp2.AddCallback(func(value float32) {
		err := hardwarePwm.SetPwm(raspi.ServoNumber(2), raspi.Pwm(value))
		if err != nil {
			log.Println(err)
		}
	})
	servoexp3.AddCallback(func(value float32) {
		err := hardwarePwm.SetPwm(raspi.ServoNumber(3), raspi.Pwm(value))
		if err != nil {
			log.Println(err)
		}
	})
	servoexp4.AddCallback(func(value float32) {
		err := hardwarePwm.SetPwm(raspi.ServoNumber(4), raspi.Pwm(value))
		if err != nil {
			log.Println(err)
		}
	})

	// Настройка mavlink соединения
	//endpointConf := gomavlib.EndpointSerial{Device: "com4", Baud: 57600} // Serial конфигурация
	endpointConf := gomavlib.EndpointTCPClient{"127.0.0.1:5600"} // TCP конфигурация
	connection := mavlink.NewConnection(
		endpointConf,
		10,
		mavlink.WithParamManager(params),
		mavlink.WithDebug(true),
	)
	err = connection.Open()
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	// Действия при обновлении параметра
	//actionsManager := actions.NewActionManager(connection)
	//ScrGoGoEnable.AddCallback(func(value float32) {
	//	fmt.Printf("GoGo %s: %.0f\n", goGoEnableParamId, value)
	//	actionsManager.GoGo(goGoEnableParamId, value)
	//})

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
