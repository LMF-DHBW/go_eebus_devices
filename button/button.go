package main // replace this with eebus

import (
	"fmt"
	"os"
	"sync"
	"time"

	eebus "github.com/LMF-DHBW/go_eebus"
	"github.com/LMF-DHBW/go_eebus/resources"

	"github.com/stianeikeland/go-rpio"
)

var (
	pin = rpio.Pin(27)
)

func main() {
	initGpio()
	// What to do for switches:
	// Connect to LEDs, receive on/off cmd -> send to all binding partners
	eebusNode := eebus.NewEebusNode("100.90.1.101", false, "button", "0001", "DHBW", "Button")
	buildDeviceModel(eebusNode)
	go eebusNode.Start()
	go readButton(eebusNode)

	/******* Infinite loop *******/
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func readButton(eebusNode *eebus.EebusNode) {
	sent := false
	for {
		time.Sleep(time.Second / 10)
		if pin.Read() == rpio.High && !sent {
			fmt.Println("switching")
			switchState(eebusNode)
			sent = true
		} else if pin.Read() == rpio.Low {
			sent = false
		}
	}
}

func initGpio() {
	err := rpio.Open()
	checkError(err)
	pin.Input()
}

func switchState(eebusNode *eebus.EebusNode) {
	for _, e := range eebusNode.SpineNode.Bindings {
		// Send with e.Conn from e.BindSubscribeEntry Address source to destination
		e.Send("write", resources.MakePayload("actuatorSwitchData", &resources.FunctionElement{
			Function: "toggle",
		}))
	}
}

func buildDeviceModel(eebusNode *eebus.EebusNode) {
	eebusNode.DeviceStructure.DeviceType = "Generic"
	eebusNode.DeviceStructure.DeviceAddress = "Switch 1"
	eebusNode.DeviceStructure.Entities = []*resources.EntityModel{
		{
			EntityType:    "Switch",
			EntityAddress: 0,
			Features: []*resources.FeatureModel{
				eebusNode.DeviceStructure.CreateNodeManagement(false),
				{
					FeatureType:      "ActuatorSwitch",
					FeatureAddress:   1,
					Role:             "client",
					Functions:        resources.ActuatorSwitch("button", "button for leds", nil),
					BindingTo:        []string{"ActuatorSwitch1"},
					MaxSubscriptions: 3,
					MaxBindings:      1,
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
