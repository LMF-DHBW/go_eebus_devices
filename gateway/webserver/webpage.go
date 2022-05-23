package webserver

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	eebus "github.com/LMF-DHBW/go_eebus"

	"github.com/LMF-DHBW/go_eebus/resources"
	"github.com/LMF-DHBW/go_eebus/ship"
)

type Node struct {
	Name    string
	Classes string
}

type Edge struct {
	Name    string
	Target  string
	Source  string
	Classes string
}

type Mainpage struct {
	ConnectedDevices string
	Nodes            []Node
	Edges            []Edge
}

type AddNewPage struct {
	Requests []Request
}

type Request struct {
	Type    string
	Brand   string
	Address string
}

type RemovePage struct {
	Trusted []Trusted
}

type Trusted struct {
	Address string
}

type HvacPage struct {
	TempX string
	TempY string
	DbErr bool
	Fan   Switchable
}

type EnergyManagementPage struct {
	ConsumptionX string
	ConsumptionY string
	SolarX       string
	SolarY       string
	BatteryX     string
	BatteryY     string
	DbErr        bool
	Relay        Switchable
}

type Measurement struct {
	Value     float64
	Timestamp string
}

type ControlSwitchablePage struct {
	Switchables []Switchable
}

type Switchable struct {
	Index      string
	Background string
	Text       string
}

func StartWebServer(eebusNode *eebus.EebusNode) {

	db := ConnectDb()

	overviewTmpl, err := template.ParseFiles("webserver/templates/index.html")
	resources.CheckError(err)
	addNewTmpl, err := template.ParseFiles("webserver/templates/addNew.html")
	resources.CheckError(err)
	removeTmpl, err := template.ParseFiles("webserver/templates/remove.html")
	resources.CheckError(err)
	lightsTmpl, err := template.ParseFiles("webserver/templates/lights.html")
	resources.CheckError(err)

	hvacTmpl, err := template.ParseFiles("webserver/templates/hvac.html")
	resources.CheckError(err)
	energyManagementTmpl, err := template.ParseFiles("webserver/templates/energy_management.html")
	resources.CheckError(err)

	_, err = os.Getwd()
	if err != nil {
		log.Println(err)
	}
	http.Handle("/stylesheets/", http.StripPrefix("/stylesheets/", http.FileServer(http.Dir("webserver/stylesheets"))))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("webserver/assets"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		Nodes := []Node{{eebusNode.DeviceStructure.DeviceAddress, "gateway"}}
		Edges := []Edge{}

		processedBindings := make([]*resources.BindSubscribeEntry, 0)
		processedSubscriptions := make([]*resources.BindSubscribeEntry, 0)

		for i, e := range eebusNode.SpineNode.Connections {
			Nodes = append(Nodes, Node{e.Address, ""})

			for j, feature := range e.SubscriptionData {
				if feature.FeatureType == "NodeManagement" {
					if feature.FunctionName == "nodeManagementBindingData" {
						var bindingData *resources.NodeManagementBindingData
						err := xml.Unmarshal([]byte(feature.CurrentState), &bindingData)
						if err == nil {
							for k, entry := range bindingData.BindingEntries {
								entryInList := false
								for x := range processedBindings {
									if entry.ClientAddress.Device == processedBindings[x].ClientAddress.Device && entry.ServerAddress.Device == processedBindings[x].ServerAddress.Device {
										entryInList = true
										break
									}
								}
								if !entryInList {
									processedBindings = append(processedBindings, entry)
									Edges = append(Edges, Edge{fmt.Sprintf("e%d%d%d", i, j, k), entry.ClientAddress.Device, entry.ServerAddress.Device, "binding"})
								}
							}
						}
					} else if feature.FunctionName == "nodeManagementSubscriptionData" {
						var subscriptionData *resources.NodeManagementSubscriptionData
						err := xml.Unmarshal([]byte(feature.CurrentState), &subscriptionData)
						if err == nil {

							for k, entry := range subscriptionData.SubscriptionEntries {
								entryInList := false
								for x := range processedSubscriptions {
									if entry.ClientAddress.Device == processedSubscriptions[x].ClientAddress.Device && entry.ServerAddress.Device == processedSubscriptions[x].ServerAddress.Device {
										entryInList = true
										break
									}
								}
								if !entryInList {
									processedSubscriptions = append(processedSubscriptions, entry)
									Edges = append(Edges, Edge{fmt.Sprintf("e%d%d%d", i, j, k), entry.ClientAddress.Device, entry.ServerAddress.Device, "subscription"})
								}
							}
						}
					}
				}
			}
		}
		overviewTmpl.Execute(w, Mainpage{
			ConnectedDevices: strconv.Itoa(len(eebusNode.SpineNode.Connections)),
			Nodes:            Nodes,
			Edges:            Edges,
		})
	})

	http.HandleFunc("/add_new", func(w http.ResponseWriter, r *http.Request) {
		acceptId, ok := r.URL.Query()["acceptId"]
		page := &AddNewPage{Requests: make([]Request, 0)}
		for i, e := range eebusNode.SpineNode.ShipNode.Requests {
			if ok && len(acceptId[0]) > 0 {
				if acceptId[0] == strconv.Itoa(i) {
					eebusNode.SpineNode.ShipNode.Requests = append(eebusNode.SpineNode.ShipNode.Requests[:i], eebusNode.SpineNode.ShipNode.Requests[i+1:]...)
					go eebusNode.SpineNode.ShipNode.Connect(e.Path, e.Ski)
					http.Redirect(w, r, "/add_new", http.StatusSeeOther)
				}
			} else {
				splittedString := strings.Split(e.Id, " ")

				page.Requests = append(page.Requests, Request{
					Type:    splittedString[0],
					Brand:   splittedString[1],
					Address: splittedString[2],
				})
			}
		}
		addNewTmpl.Execute(w, page)

	})

	http.HandleFunc("/remove", func(w http.ResponseWriter, r *http.Request) {
		removeId, ok := r.URL.Query()["removeId"]

		page := &RemovePage{Trusted: make([]Trusted, 0)}
		skis, devices := ship.ReadSkis()
		for i := range skis {
			if ok && len(removeId[0]) > 0 {
				if removeId[0] == strconv.Itoa(i) {
					skis = append(skis[:i], skis[i+1:]...)
					devices = append(devices[:i], devices[i+1:]...)
					ship.WriteSkis(skis, devices)
					http.Redirect(w, r, "/remove", http.StatusSeeOther)
				}
			} else {
				page.Trusted = append(page.Trusted, Trusted{
					Address: devices[i],
				})
			}
		}
		removeTmpl.Execute(w, page)
	})

	http.HandleFunc("/lights", func(w http.ResponseWriter, r *http.Request) {

		ledId, ok := r.URL.Query()["ledId"]
		page := &ControlSwitchablePage{Switchables: make([]Switchable, 0)}
		fnum := 0
		for _, e := range eebusNode.SpineNode.Connections {
			for _, feature := range e.SubscriptionData {
				if strings.HasPrefix(feature.FeatureType, "ActuatorSwitch") && feature.EntityType == "LED" {
					var state *resources.FunctionElement
					err := xml.Unmarshal([]byte(feature.CurrentState), &state)
					if err == nil {
						if ok && len(ledId[0]) > 0 {
							// Switch LED on/off
							if ledId[0] == strconv.Itoa(fnum+1) {
								// Send request to LED to switch state
								e.SendXML(e.OwnDevice.MakeHeader(0, 0, &feature.FeatureAddress, "write", e.MsgCounter, false), resources.MakePayload("actuatorSwitchData", &resources.FunctionElement{
									Function: "toggle",
								}))
								time.Sleep(time.Second / 5)
								http.Redirect(w, r, "/lights", http.StatusSeeOther)
								//eebusNode.SpineNode.ShipNode.Requests = append(eebusNode.SpineNode.ShipNode.Requests[:i], eebusNode.SpineNode.ShipNode.Requests[i+1:]...)
							}
						} else {
							// Get all LEDs --> subscribe to LEDs and save their state
							background := "danger"
							text := "Turn off"
							if state.Function == "off" {
								background = "success"
								text = "Turn on"
							}
							page.Switchables = append(page.Switchables, Switchable{
								Background: background,
								Text:       text,
								Index:      strconv.Itoa(fnum + 1),
							})
						}
					}
					fnum++
				}
			}
		}

		lightsTmpl.Execute(w, page)
	})

	http.HandleFunc("/hvac", func(w http.ResponseWriter, r *http.Request) {
		vals, time := make([]string, 0), make([]string, 0)
		dbErr := false
		if db != nil {
			results, err := db.Query("SELECT value, timestamp FROM temperature WHERE timestamp > DATE(NOW() - INTERVAL 1 DAY) ORDER BY ID DESC")
			if err == nil {
				vals, time = getMeasurementValues(results)
			} else {
				dbErr = true
			}
		} else {
			dbErr = true
		}

		fanId, ok := r.URL.Query()["fanId"]
		fan := Switchable{}
		fnum := 0
		for _, e := range eebusNode.SpineNode.Connections {
			for _, feature := range e.SubscriptionData {
				if feature.FeatureType == "ActuatorSwitch" && feature.EntityType == "Fan" {
					var state *resources.FunctionElement
					err := xml.Unmarshal([]byte(feature.CurrentState), &state)
					if err == nil {
						if ok && len(fanId[0]) > 0 {
							// Switch Fan on/off
							if fanId[0] == strconv.Itoa(fnum) {
								// Send request to Fan to switch state
								e.SendXML(e.OwnDevice.MakeHeader(0, 0, &feature.FeatureAddress, "write", e.MsgCounter, false), resources.MakePayload("actuatorSwitchData", &resources.FunctionElement{
									Function: "toggle",
								}))
								http.Redirect(w, r, "/hvac", http.StatusSeeOther)
							}
						} else {
							background := "green"
							text := "Turn off"
							if state.Function == "off" {
								background = "red"
								text = "Turn on"
							}
							fan = Switchable{
								Background: background,
								Text:       text,
								Index:      strconv.Itoa(fnum),
							}
						}
					}
					fnum++
				}
			}
		}

		hvacTmpl.Execute(w, &HvacPage{strings.Join(time, ","), strings.Join(vals, ","), dbErr, fan})
	})

	http.HandleFunc("/energy-management", func(w http.ResponseWriter, r *http.Request) {
		solarY, solarX := make([]string, 0), make([]string, 0)
		consumptionY, consumptionX := make([]string, 0), make([]string, 0)
		batteryY, batteryX := make([]string, 0), make([]string, 0)
		dbErr := false
		if db != nil {
			results, err := db.Query("SELECT value, timestamp FROM solar WHERE timestamp > DATE_SUB(CURDATE(), INTERVAL 1 DAY) ORDER BY ID DESC")
			if err == nil {
				solarY, solarX = getMeasurementValues(results)
			} else {
				dbErr = true
			}
			results, err = db.Query("SELECT value, timestamp FROM consumption WHERE timestamp > DATE_SUB(CURDATE(), INTERVAL 1 DAY) ORDER BY ID DESC")
			if err == nil {
				consumptionY, consumptionX = getMeasurementValues(results)
			} else {
				dbErr = true
			}
			results, err = db.Query("SELECT value, timestamp FROM battery WHERE timestamp > DATE_SUB(CURDATE(), INTERVAL 1 DAY) ORDER BY ID DESC")
			if err == nil {
				batteryY, batteryX = getMeasurementValues(results)
			} else {
				dbErr = true
			}
		} else {
			dbErr = true
		}

		relayId, ok := r.URL.Query()["relayId"]
		relay := Switchable{}
		fnum := 0
		for _, e := range eebusNode.SpineNode.Connections {
			for _, feature := range e.SubscriptionData {
				if feature.FeatureType == "ActuatorSwitch" && feature.EntityType == "Relay" {
					var state *resources.FunctionElement
					err := xml.Unmarshal([]byte(feature.CurrentState), &state)
					if err == nil {
						if ok && len(relayId[0]) > 0 {
							// Switch relay on/off
							if relayId[0] == strconv.Itoa(fnum) {
								// Send request to relay to switch state
								e.SendXML(e.OwnDevice.MakeHeader(0, 0, &feature.FeatureAddress, "write", e.MsgCounter, false), resources.MakePayload("actuatorSwitchData", &resources.FunctionElement{
									Function: "toggle",
								}))
								http.Redirect(w, r, "/energy-management", http.StatusSeeOther)
							}
						} else {
							background := "#5DBE7C"
							text := "Switch to solar power"
							if state.Function == "off" {
								background = "#8B1C00"
								text = "Switch to socket power"
							}
							relay = Switchable{
								Background: background,
								Text:       text,
								Index:      strconv.Itoa(fnum),
							}
						}
					}
					fnum++
				}
			}
		}

		energyManagementTmpl.Execute(w, &EnergyManagementPage{
			strings.Join(consumptionX, ","), strings.Join(consumptionY, ","),
			strings.Join(solarX, ","), strings.Join(solarY, ","),
			strings.Join(batteryX, ","), strings.Join(batteryY, ","), dbErr, relay})
	})

	err = http.ListenAndServe(":8081", nil)
	resources.CheckError(err)
}

func getMeasurementValues(results *sql.Rows) ([]string, []string) {
	vals := make([]string, 0)
	time := make([]string, 0)

	for results.Next() {
		var measurement Measurement
		// for each row, scan the result into our tag composite object
		err := results.Scan(&measurement.Value, &measurement.Timestamp)
		if err == nil {
			// and then print out the tag's Name attribute
			vals = append(vals, fmt.Sprintf("%f", measurement.Value))
			time = append(time, "\""+measurement.Timestamp+"\"")
		}
	}
	return vals, time
}
