#!/bin/bash

# Проверка на root права
if [ "$EUID" -ne 0 ]; then
    echo "Требуются root права"
    exit 1
fi

echo "Установка брокера"
sudo apt install -y mosquitto mosquitto-clients

echo "Активация сервиса"
sudo systemctl enable mosquitto.service

echo "Создание конфиг файла для разрешения работы по сети"
cat > /etc/mosquitto/conf.d/network.conf << EOF
listener 1883
allow_anonymous true
EOF

echo "Перезапуск сервиса"
sudo systemctl restart mosquitto
