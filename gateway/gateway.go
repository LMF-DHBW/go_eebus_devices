package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"strings"
	"sync"

	"go_eebus_gateway/webserver"

	eebus "github.com/LMF-DHBW/go_eebus"
	"github.com/LMF-DHBW/go_eebus/resources"
	"github.com/LMF-DHBW/go_eebus/spine"
)

var db *sql.DB

func main() {

	// What to do for switches:
	// Connect to LEDs, receive on/off cmd -> send to all binding partners
	eebusNode := eebus.NewEebusNode("100.90.1.109", true, "gateway", "0001", "DHBW", "Gateway")
	eebusNode.Update = update
	buildDeviceModelGateway(eebusNode)

	go eebusNode.Start()
	go webserver.StartWebServer(eebusNode)

	db = webserver.ConnectDb()
	/******* Infinite loop *******/
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func update(datagram resources.DatagramType, conn spine.SpineConnection) {
	// check from where the update came: temp sensor, current sensor, LED, save new state
	entitySource := datagram.Header.AddressSource.Entity
	featureSource := datagram.Header.AddressSource.Feature

	if strings.HasPrefix(conn.DiscoveryInformation.FeatureInformation[featureSource].Description.FeatureType, "Measurement") {
		table := ""
		if conn.DiscoveryInformation.EntityInformation[entitySource].Description.EntityType == "Temperature" {
			table = "temperature"
		} else if conn.DiscoveryInformation.EntityInformation[entitySource].Description.EntityType == "Solar" {
			table = "solar"
		} else if conn.DiscoveryInformation.EntityInformation[entitySource].Description.EntityType == "Consumption" {
			table = "consumption"
		} else if conn.DiscoveryInformation.EntityInformation[entitySource].Description.EntityType == "Battery" {
			table = "battery"
		}

		var Function *resources.MeasurementDataType
		err := xml.Unmarshal([]byte(datagram.Payload.Cmd.Function), &Function)
		if err == nil {
			_, err = db.Query("INSERT INTO " + table + " (`value`) VALUES (" + fmt.Sprintf("%f", Function.Value) + ")")
			if err != nil {
				log.Println(err)
				log.Println("Cant insert")
			}
		}
	}
}

func buildDeviceModelGateway(eebusNode *eebus.EebusNode) {
	eebusNode.DeviceStructure.DeviceType = "Gateway"
	eebusNode.DeviceStructure.DeviceAddress = "Gateway"
	eebusNode.DeviceStructure.Entities = []*resources.EntityModel{
		{
			EntityType:    "Gateway",
			EntityAddress: 0,
			Features: []*resources.FeatureModel{
				eebusNode.DeviceStructure.CreateNodeManagement(true),
			},
		},
	}
}
