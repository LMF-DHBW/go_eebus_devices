package main

import (
	"log"
	"math"
	"sync"
	"time"

	eebus "github.com/LMF-DHBW/go_eebus"
	"github.com/LMF-DHBW/go_eebus/resources"

	"golang.org/x/exp/io/i2c"
)

func main() {
	// What to do for temperature:
	// Send temp to subscription partners
	eebusNode := eebus.NewEebusNode("100.90.1.102", false, "temp", "0001", "DHBW", "Temperature sensor")
	buildDeviceModel(eebusNode)
	eebusNode.Start()

	readTemp(eebusNode)
	/******* Infinite loop *******/
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func readTemp(eebusNode *eebus.EebusNode) {
	var prevTemp float64 = -1
	for {
		temp := readI2cTemp()
		if math.Abs(temp-prevTemp) > 0 || true {
			for _, e := range eebusNode.SpineNode.Subscriptions {
				e.Send("notify", resources.MakePayload("measurementData", resources.MeasurementDataType{
					ValueType:   "value",
					Value:       float64(temp),
					ValueSource: "measuredValue",
					ValueState:  "normal",
				}))
			}
		}
		prevTemp = temp
		time.Sleep(time.Second * 5)
	}
}

func readI2cTemp() float64 {
	i2c, err := i2c.Open(&i2c.Devfs{Dev: "/dev/i2c-1"}, 0x38)
	if err != nil {
		log.Fatal(err)
	}

	defer i2c.Close()
	err = i2c.Write([]byte{0xE1, 0x08, 0x00})
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second / 2)

	err = i2c.Read(make([]byte, 1))
	if err != nil {
		log.Fatal(err)
	}

	err = i2c.Write([]byte{0xAC, 0x33, 0x00})
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second / 2)

	data := make([]byte, 15)
	err = i2c.Read(data)
	if err != nil {
		log.Fatal(err)
	}

	t := (int(data[3]&0x0F) << 16) | (int(data[4]) << 8) | int(data[5])
	temp := ((float64(t) * 200) / 1048576) - 50
	log.Println("Temp: ", temp)

	h := ((int(data[1]) << 16) | (int(data[2]) << 8) | int(data[3])) >> 4
	humid := int(h) * 100 / 1048576
	log.Println("Humid: ", humid)
	return temp
}

func buildDeviceModel(eebusNode *eebus.EebusNode) {
	eebusNode.DeviceStructure.DeviceType = "Generic"
	eebusNode.DeviceStructure.DeviceAddress = "Temperature"
	eebusNode.DeviceStructure.Entities = []*resources.EntityModel{
		{
			EntityType:    "Temperature",
			EntityAddress: 0,
			Features: []*resources.FeatureModel{
				eebusNode.DeviceStructure.CreateNodeManagement(false),
				{
					FeatureType:      "MeasurementTemp",
					FeatureAddress:   1,
					Functions:        resources.Measurement("temperature", "degC", "roomAirTemperature", "temp", "measurement of temperature"),
					MaxSubscriptions: 128,
					Role:             "server",
				},
			},
		},
	}
	eebusNode.DeviceStructure.Entities[0].Features[0] = eebusNode.DeviceStructure.CreateNodeManagement(false)
}
