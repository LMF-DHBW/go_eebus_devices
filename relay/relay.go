package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/LMF-DHBW/go_eebus/resources"

	eebus "github.com/LMF-DHBW/go_eebus"
	"github.com/LMF-DHBW/go_eebus/spine"

	"github.com/stianeikeland/go-rpio"
)

var eebusNode *eebus.EebusNode
var solar float64
var battery float64
var (
	pin = rpio.Pin(17)
)

func main() {
	// What to do for leds:
	// Connect to switches, receive SPINE on/off requests
	initGpio()
	eebusNode = eebus.NewEebusNode("100.90.1.109", false, "relay", "0001", "DHBW", "Relay")
	buildDeviceModel(eebusNode)
	eebusNode.Update = update
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

func switchRelay(funcName string, in string, featureAddr resources.FeatureAddressType) {
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
	eebusNode.DeviceStructure.DeviceAddress = "Relay"
	eebusNode.DeviceStructure.Entities = []*resources.EntityModel{
		{
			EntityType:    "Relay",
			EntityAddress: 0,
			Features: []*resources.FeatureModel{
				eebusNode.DeviceStructure.CreateNodeManagement(false),
				{
					FeatureType:      "ActuatorSwitch",
					FeatureAddress:   1,
					Role:             "server",
					Functions:        resources.ActuatorSwitch("relay", "relay that works as energy switch", switchRelay),
					SubscriptionTo:   []string{"MeasurementSolar", "MeasurementBattery"},
					MaxSubscriptions: 3,
					MaxBindings:      3,
				},
			},
		},
	}
	eebusNode.DeviceStructure.Entities[0].Features[0] = eebusNode.DeviceStructure.CreateNodeManagement(false)
}

func update(datagram resources.DatagramType, conn spine.SpineConnection) {
	featureSource := datagram.Header.AddressSource.Feature
	featureType := conn.DiscoveryInformation.FeatureInformation[featureSource].Description.FeatureType
	if strings.HasPrefix(featureType, "Measurement") {
		var Function *resources.MeasurementDataType
		err := xml.Unmarshal([]byte(datagram.Payload.Cmd.Function), &Function)
		if err == nil {
			if featureType == "MeasurementSolar" {
				solar = Function.Value
			} else if featureType == "MeasurementBattery" {
				battery = Function.Value
			}
			if battery != 0 {
				if battery < 11.2 {
					pin.High()
				}
			}
		}
	}
}

/************ Helper functions ************/

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
