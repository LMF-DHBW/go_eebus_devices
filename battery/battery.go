package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	eebus "github.com/LMF-DHBW/go_eebus"
	"github.com/LMF-DHBW/go_eebus/resources"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

func main() {
	eebusNode := eebus.NewEebusNode("100.90.1.109", false, "battery", "0001", "DHBW", "Battery")
	buildDeviceModel(eebusNode)
	eebusNode.Start()

	r := raspi.NewAdaptor()
	ads := i2c.NewADS1015Driver(r, i2c.WithBus(1), i2c.WithAddress(0x48))

	work := func() {
		gobot.Every(5*time.Second, func() {
			volt, _ := ads.ReadWithDefaults(0)
			volt = float64(volt) * 5.546153
			log.Println(volt, "V")
			if volt > 9 {
				for _, e := range eebusNode.SpineNode.Subscriptions {
					e.Send("notify", resources.MakePayload("measurementData", resources.MeasurementDataType{
						ValueType:   "value",
						Value:       volt,
						ValueSource: "measuredValue",
						ValueState:  "normal",
					}))
				}
			}
		})
	}

	robot := gobot.NewRobot("adsbot",
		[]gobot.Connection{r},
		[]gobot.Device{ads},
		work,
	)

	robot.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			os.Exit(0)
		}
	}()

	/******* Infinite loop *******/
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func buildDeviceModel(eebusNode *eebus.EebusNode) {
	eebusNode.DeviceStructure.DeviceType = "Generic"
	eebusNode.DeviceStructure.DeviceAddress = "Battery"
	eebusNode.DeviceStructure.Entities = []*resources.EntityModel{
		{
			EntityType:    "Battery",
			EntityAddress: 0,
			Features: []*resources.FeatureModel{
				eebusNode.DeviceStructure.CreateNodeManagement(false),
				{
					FeatureType:      "MeasurementBattery",
					FeatureAddress:   1,
					Functions:        resources.Measurement("voltage", "V", "acYieldTotal", "volt", "measurement of voltage"),
					MaxSubscriptions: 128,
					Role:             "server",
				},
			},
		},
	}
	eebusNode.DeviceStructure.Entities[0].Features[0] = eebusNode.DeviceStructure.CreateNodeManagement(false)
}
