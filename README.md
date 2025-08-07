# gomavclient
Client of mavlink

Для работы может потребоваться брокер сообщений mosquitto https://mosquitto.org/download/

## Для компиляции под Linux на Windows:
1. Узнать архитектуру системы linux:
```shell
uname -m
```
2. Скомпилировать для этой ОС и архитектуры в Windows PowerShell:
```shell
$env:GOOS = "linux"
$env:GOARCH = "arm64"
$env:CGO_ENABLED = "0"
go build -o gomavclient
```
Скопировать с windows на linux:
```shell
pscp.exe gomavclient pi@<ip_address_of_raspberry_pi>:/home/pi/
chmod +x gomavclient
```
# Установка сервиса
1. Создайте файл службы:
```shell
sudo nano /etc/systemd/system/gomavclient.service
```
2. Добавьте следующее содержимое:
```shell
[Unit]
Description=GoMavClient Service
After=network.target

[Service]
ExecStart=/home/pi/gomavclient/gomavclient
WorkingDirectory=/home/pi/gomavclient
User=pi
Group=pi
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
