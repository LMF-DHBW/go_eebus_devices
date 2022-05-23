package main

import (
	"log"
	"sync"
	"time"

	eebus "github.com/LMF-DHBW/go_eebus"
	ina219 "github.com/NeuralSpaz/ti-ina219"

	"github.com/LMF-DHBW/go_eebus/resources"
)

func main() {
	eebusNode := eebus.NewEebusNode("100.90.1.109", false, "solar", "0001", "DHBW", "Solar panel")
	buildDeviceModel(eebusNode)
	eebusNode.Start()

	readPower(eebusNode)
	/******* Infinite loop *******/
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func readPower(eebusNode *eebus.EebusNode) {
	for {
		power := readI2cPower()
		for _, e := range eebusNode.SpineNode.Subscriptions {
			e.Send("notify", resources.MakePayload("measurementData", resources.MeasurementDataType{
				ValueType:   "value",
				Value:       float64(power),
				ValueSource: "measuredValue",
				ValueState:  "normal",
			}))
		}
		time.Sleep(time.Second * 5)
	}
}

func readI2cPower() float64 {
	sensor := ina219.New(0x40, 1)
	if err := ina219.Fetch(sensor); err != nil {
		log.Println(err)
	}
	log.Println(sensor.Current*sensor.Power, "W")
	return sensor.Current * sensor.Power
}

func buildDeviceModel(eebusNode *eebus.EebusNode) {
	eebusNode.DeviceStructure.DeviceType = "Generic"
	eebusNode.DeviceStructure.DeviceAddress = "Solar Panels"
	eebusNode.DeviceStructure.Entities = []*resources.EntityModel{
		{
			EntityType:    "Solar",
			EntityAddress: 0,
			Features: []*resources.FeatureModel{
				eebusNode.DeviceStructure.CreateNodeManagement(false),
				{
					FeatureType:      "MeasurementSolar",
					FeatureAddress:   1,
					Functions:        resources.Measurement("power", "W", "acPowerTotal", "pow", "measurement of power"),
					MaxSubscriptions: 128,
					Role:             "server",
				},
			},
		},
	}
	eebusNode.DeviceStructure.Entities[0].Features[0] = eebusNode.DeviceStructure.CreateNodeManagement(false)
}
