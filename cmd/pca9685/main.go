package main

import (
	"log"
	"time"

	"github.com/googolgl/go-i2c"
	"github.com/googolgl/go-pca9685"
)

func main() {
	// Создаем подключение к I2C-шине (адрес PCA9685 по умолчанию - 0x40)
	i2c, err := i2c.New(pca9685.Address, "/dev/i2c-1") // Убедитесь, что используете правильный путь для вашей Raspberry Pi
	if err != nil {
		log.Fatal(err)
	}

	// Инициализируем драйвер PCA9685
	pca, err := pca9685.New(i2c, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Устанавливаем частоту PWM (например, 50 Гц для сервоприводов)
	err = pca.SetFreq(50)
	if err != nil {
		log.Fatal(err)
	}

	// Создаем объект для управления сервоприводом на канале 0
	servo := pca.ServoNew(0, nil)

	// Пошаговое изменение угла от 0° до 180°
	for angle := 0; angle <= 180; angle += 10 {
		err = servo.Angle(angle)
		if err != nil {
			log.Printf("Ошибка установки угла: %v", err)
			continue
		}
		time.Sleep(500 * time.Millisecond) // Задержка для плавного движения
	}

	log.Println("Управление завершено!")
}
