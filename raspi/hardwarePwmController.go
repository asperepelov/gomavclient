package raspi

import (
	"fmt"
	"github.com/stianeikeland/go-rpio/v4"
	"sync"
)

// HardwarePWMController управляет аппаратным PWM на Raspberry Pi
type HardwarePWMController struct {
	pin       rpio.Pin
	minValue  int // минимальное значение (1000)
	maxValue  int // максимальное значение (2000)
	value     int // текущее значение
	mu        sync.Mutex
	isStarted bool
}

// NewHardwarePWMController создает новый контроллер аппаратного PWM
func NewHardwarePWMController() (*HardwarePWMController, error) {
	// Открываем GPIO
	if err := rpio.Open(); err != nil {
		return nil, fmt.Errorf("failed to open GPIO: %v", err)
	}

	// GPIO 13 соответствует PWM1 на RPi4
	pin := rpio.Pin(13)

	// Устанавливаем режим PWM
	pin.Mode(rpio.Pwm)

	// Устанавливаем частоту PWM (50Hz - стандарт для сервоприводов)
	// Clock делитель для RPi: 19.2MHz / 384 = 50kHz
	// Диапазон 1000 даст нам 50Hz (50kHz / 1000 = 50Hz)
	rpio.PwmSetClockDivider(384)
	pin.PwmSetRange(1000)

	controller := &HardwarePWMController{
		pin:       pin,
		minValue:  1000,
		maxValue:  2000,
		value:     1500, // значение по умолчанию (среднее)
		isStarted: false,
	}

	// Устанавливаем начальное значение PWM (0 для начала)
	pin.PwmSetData(0)

	return controller, nil
}

// SetValue устанавливает значение PWM от 1000 до 2000
func (pwm *HardwarePWMController) SetValue(value int) error {
	pwm.mu.Lock()
	defer pwm.mu.Unlock()

	// Проверка диапазона
	if value < pwm.minValue || value > pwm.maxValue {
		return fmt.Errorf("value must be between %d and %d, got %d",
			pwm.minValue, pwm.maxValue, value)
	}

	pwm.value = value

	// Если контроллер запущен, обновляем значение на пине
	if pwm.isStarted {
		// Преобразуем значение из диапазона 1000-2000 в 50-100
		// 1000 соответствует 5% (50 из 1000) рабочего цикла, 2000 - 10% (100 из 1000)
		// Это стандартные значения для сервоприводов (1-2мс от 20мс периода)
		pwmData := (value-pwm.minValue)/20 + 50
		pwm.pin.PwmSetData(uint32(pwmData))
	}

	return nil
}

// GetValue возвращает текущее значение PWM
func (pwm *HardwarePWMController) GetValue() int {
	pwm.mu.Lock()
	defer pwm.mu.Unlock()

	return pwm.value
}

// Start запускает генерацию PWM сигнала
func (pwm *HardwarePWMController) Start() error {
	pwm.mu.Lock()
	defer pwm.mu.Unlock()

	if pwm.isStarted {
		return nil // уже запущен
	}

	// Устанавливаем текущее значение
	pwmData := (pwm.value-pwm.minValue)/20 + 50
	pwm.pin.PwmSetData(uint32(pwmData))

	pwm.isStarted = true
	return nil
}

// Stop останавливает генерацию PWM сигнала
func (pwm *HardwarePWMController) Stop() {
	pwm.mu.Lock()
	defer pwm.mu.Unlock()

	if !pwm.isStarted {
		return // уже остановлен
	}

	// Устанавливаем 0 для остановки
	pwm.pin.PwmSetData(0)
	pwm.isStarted = false
}

// Close освобождает ресурсы
func (pwm *HardwarePWMController) Close() {
	pwm.Stop()
	rpio.Close()
}
