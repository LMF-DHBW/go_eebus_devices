package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"sync"

	"github.com/LMF-DHBW/go_eebus/resources"

	eebus "github.com/LMF-DHBW/go_eebus"

	"github.com/stianeikeland/go-rpio"
)

var eebusNode *eebus.EebusNode
var (
	pin = rpio.Pin(17)
)

func main() {
	initGpio()
	eebusNode = eebus.NewEebusNode("100.90.1.99", false, "led2", "0002", "DHBW", "LED")
	buildDeviceModel(eebusNode)
	eebusNode.Start()

	/******* Infinite loop *******/
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
	defer rpio.Close()
}

func initGpio() {
	err := rpio.Open()
	checkError(err)
	pin.Output()
}

func switchLed(funcName string, in string, featureAddr resources.FeatureAddressType) {
	var Function *resources.FunctionElement

	if funcName == "actuatorSwitchData" {
		err := xml.Unmarshal([]byte(in), &Function)
		if err == nil {
			feature := eebusNode.DeviceStructure.Entities[featureAddr.Entity].Features[featureAddr.Feature]

			for i := range feature.Functions {
				if feature.Functions[i].FunctionName == funcName {
					if Function.Function == "toggle" {
						currentState := eebusNode.DeviceStructure.Entities[featureAddr.Entity].Features[featureAddr.Feature].Functions[i].Function.(*resources.FunctionElement).Function
						if currentState == "on" {
							eebusNode.DeviceStructure.Entities[featureAddr.Entity].Features[featureAddr.Feature].Functions[i].Function.(*resources.FunctionElement).Function = "off"
						} else {
							eebusNode.DeviceStructure.Entities[featureAddr.Entity].Features[featureAddr.Feature].Functions[i].Function.(*resources.FunctionElement).Function = "on"
						}
					} else {
						eebusNode.DeviceStructure.Entities[featureAddr.Entity].Features[featureAddr.Feature].Functions[i].Function = Function
					}
					currentState := eebusNode.DeviceStructure.Entities[featureAddr.Entity].Features[featureAddr.Feature].Functions[i].Function.(*resources.FunctionElement).Function
					if currentState == "on" {
						pin.High()
					} else {
						pin.Low()
					}
					break
				}
			}

			for _, e := range eebusNode.SpineNode.Subscriptions {
				// Send with e.Conn from e.BindSubscribeEntry Address source to destination
				e.Conn.SendXML(
					e.Conn.OwnDevice.MakeHeader(e.BindSubscribeEntry.ServerAddress.Entity, e.BindSubscribeEntry.ServerAddress.Feature, resources.MakeFeatureAddress(e.BindSubscribeEntry.ClientAddress.Device, e.BindSubscribeEntry.ClientAddress.Entity, e.BindSubscribeEntry.ClientAddress.Feature), "notify", e.Conn.MsgCounter, false),
					resources.MakePayload("actuatorSwitchData", eebusNode.DeviceStructure.Entities[0].Features[1].Functions[0].Function))
			}
		}
	}
}

func buildDeviceModel(eebusNode *eebus.EebusNode) {
	eebusNode.DeviceStructure.DeviceType = "Generic"
	eebusNode.DeviceStructure.DeviceAddress = "LED 2"
	eebusNode.DeviceStructure.Entities = []*resources.EntityModel{
		{
			EntityType:    "LED",
			EntityAddress: 0,
			Features: []*resources.FeatureModel{
				eebusNode.DeviceStructure.CreateNodeManagement(false),
				{
					FeatureType:      "ActuatorSwitch2",
					FeatureAddress:   1,
					Role:             "server",
					Functions:        resources.ActuatorSwitch("switch", "generic switch", switchLed),
					MaxSubscriptions: 3,
					MaxBindings:      3,
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
