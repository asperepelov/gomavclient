package raspi

import (
	"fmt"
	"log"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
	//"periph.io/x/periph/conn/gpio/gpioreg"
)

type Pwm int32 // Скважность
type Pin struct {
	Name      string
	InitValue Pwm
}

// HardwarePWMController управляет аппаратным PWM
type HardwarePWMController struct {
	Pins []Pin
}

// NewHardwarePWMController создает новый контроллер аппаратного PWM
func NewHardwarePWMController(pins []Pin) (*HardwarePWMController, error) {
	if len(pins) == 0 {
		return nil, fmt.Errorf("no pins specified for hardware PWM")
	}

	hc := &HardwarePWMController{
		Pins: pins,
	}
	err := hc.init()
	if err != nil {
		return nil, fmt.Errorf("failed to init hardware PWM controller: %w", err)
	}

	return hc, nil
}

func (hc *HardwarePWMController) init() error {
	_, err := host.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize host: %w", err)
	}

	// Set PWM
	for i, pin := range hc.Pins {
		err = hc.SetPwm(ServoNumber(i+1), pin.InitValue)
		if err != nil {
			return fmt.Errorf("failed to set PWM to pin %s: %w", pin.Name, err)
		}
	}

	log.Println("Hardware PWM controller initialized")
	return nil
}

type ServoNumber uint8 // Номер сервопривода начиная с 1
// SetPwm устанавливает значение PWM на выходе
func (hc *HardwarePWMController) SetPwm(num ServoNumber, value Pwm) error {
	if num < 1 || num > ServoNumber(len(hc.Pins)) {
		return fmt.Errorf("servo number out of range")
	}

	pinName := hc.Pins[num-1].Name
	pin := gpioreg.ByName(pinName)
	if pin == nil {
		return fmt.Errorf("failed to find pin %s", pinName)
	}
	duty := gpio.DutyMax * 1 / gpio.Duty(value)             // Заполнение = 1 / Скважность
	err := pin.PWM(duty, physic.Frequency(50*physic.Hertz)) // 50 Гц стандарт для сервоприводов
	if err != nil {
		return fmt.Errorf("failed to set PWM: %w", err)
	}
	return nil
}
