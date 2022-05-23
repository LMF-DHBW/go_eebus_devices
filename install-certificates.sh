#!/bin/bash
sudo cp gateway/gateway.crt /usr/local/share/ca-certificates/gateway.crt
sudo cp button/button.crt /usr/local/share/ca-certificates/button.crt
sudo cp consumption/consumption.crt /usr/local/share/ca-certificates/consumption.crt
sudo cp led/led.crt /usr/local/share/ca-certificates/led.crt
sudo cp relay/relay.crt /usr/local/share/ca-certificates/relay.crt
sudo cp solar/solar.crt /usr/local/share/ca-certificates/solar.crt
sudo cp temperature/temp.crt /usr/local/share/ca-certificates/temp.crt
sudo cp led2/led2.crt /usr/local/share/ca-certificates/led2.crt
sudo cp button2/button2.crt /usr/local/share/ca-certificates/button2.crt
sudo cp battery_display/battery_display.crt /usr/local/share/ca-certificates/battery_display.crt
sudo cp battery/battery.crt /usr/local/share/ca-certificates/battery.crt
sudo cp fan/fan.crt /usr/local/share/ca-certificates/fan.crt

update-ca-certificates
