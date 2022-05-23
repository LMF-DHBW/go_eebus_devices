package main

import (
	"log"
	"math"
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
	eebusNode := eebus.NewEebusNode("100.90.1.109", false, "consumption", "0001", "DHBW", "Consumption")
	buildDeviceModel(eebusNode)
	eebusNode.Start()

	r := raspi.NewAdaptor()
	ads := i2c.NewADS1015Driver(r, i2c.WithBus(1), i2c.WithAddress(0x48))

	work := func() {
		gobot.Every(5*time.Second, func() {
			power, _ := ads.ReadWithDefaults(1)
			power = (math.Abs(float64(power)-1.64805) / 0.04) * 1000
			log.Println(power, "mA")
			if power != 0 {
				for _, e := range eebusNode.SpineNode.Subscriptions {
					e.Send("notify", resources.MakePayload("measurementData", resources.MeasurementDataType{
						ValueType:   "value",
						Value:       (power * 5) / 1000,
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
	eebusNode.DeviceStructure.DeviceAddress = "Energy Consumption"
	eebusNode.DeviceStructure.Entities = []*resources.EntityModel{
		{
			EntityType:    "Consumption",
			EntityAddress: 0,
			Features: []*resources.FeatureModel{
				eebusNode.DeviceStructure.CreateNodeManagement(false),
				{
					FeatureType:      "Measurement",
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
