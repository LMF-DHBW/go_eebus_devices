package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"sync"

	"github.com/LMF-DHBW/go_eebus/resources"
	"github.com/LMF-DHBW/go_eebus/spine"

	eebus "github.com/LMF-DHBW/go_eebus"

	"github.com/stianeikeland/go-rpio"
)

var eebusNode *eebus.EebusNode
var (
	pin1 = rpio.Pin(5)
	pin2 = rpio.Pin(6)
	pin3 = rpio.Pin(13)
	pin4 = rpio.Pin(19)
)

func main() {
	// What to do for leds:
	// Connect to switches, receive SPINE on/off requests
	initGpio()
	eebusNode = eebus.NewEebusNode("100.90.1.101", false, "battery_display", "0001", "DHBW", "Battery display")
	eebusNode.Update = update
	buildDeviceModel(eebusNode)
	eebusNode.Start()

	//go switchLED(eebusNode)

	/******* Infinite loop *******/
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
	defer rpio.Close()
}

func initGpio() {
	err := rpio.Open()
	checkError(err)
	pin1.Output()
	pin2.Output()
	pin3.Output()
	pin4.Output()
}

func update(datagram resources.DatagramType, conn spine.SpineConnection) {
	featureSource := datagram.Header.AddressSource.Feature
	if conn.DiscoveryInformation.FeatureInformation[featureSource].Description.FeatureType == "MeasurementBattery" {
		var Function *resources.MeasurementDataType
		err := xml.Unmarshal([]byte(datagram.Payload.Cmd.Function), &Function)
		if err == nil {
			if Function.Value < 11.1 {
				pin1.Low()
				pin2.Low()
				pin3.Low()
				pin4.Low()
			} else if Function.Value >= 12.2 {
				pin1.High()
				pin2.High()
				pin3.High()
				pin4.High()
			} else if Function.Value >= 11.85 {
				pin1.High()
				pin2.High()
				pin3.High()
				pin4.Low()
			} else if Function.Value >= 11.55 {
				pin1.High()
				pin2.High()
				pin3.Low()
				pin4.Low()
			} else if Function.Value >= 11.1 {
				pin1.High()
				pin2.Low()
				pin3.Low()
				pin4.Low()
			}
		}
	}
}

func buildDeviceModel(eebusNode *eebus.EebusNode) {
	eebusNode.DeviceStructure.DeviceType = "Generic"
	eebusNode.DeviceStructure.DeviceAddress = "Battery display"
	eebusNode.DeviceStructure.Entities = []*resources.EntityModel{
		{
			EntityType:    "Battery display",
			EntityAddress: 0,
			Features: []*resources.FeatureModel{
				eebusNode.DeviceStructure.CreateNodeManagement(false),
				{
					FeatureType:      "ActuatorSwitch",
					FeatureAddress:   1,
					Role:             "server",
					Functions:        resources.ActuatorSwitch("switch", "generic switch", nil),
					SubscriptionTo:   []string{"MeasurementBattery"},
					MaxSubscriptions: 1,
				},
			},
		},
	}
	eebusNode.DeviceStructure.Entities[0].Features[0] = eebusNode.DeviceStructure.CreateNodeManagement(false)
}

/************ Helper functions ************/

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
