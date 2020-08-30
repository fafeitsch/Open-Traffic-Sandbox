package main

import (
	"fmt"
	channels2 "github.com/fafeitsch/Open-Traffic-Sandbox/channels"
	"github.com/fafeitsch/Open-Traffic-Sandbox/definition"
	"github.com/fafeitsch/Open-Traffic-Sandbox/routing"
	"github.com/fafeitsch/Open-Traffic-Sandbox/server"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("missing scenario definition and stop definition file")
	}
	scenarioFile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("could not read input file: %v", err)
	}
	defer func() { _ = scenarioFile.Close() }()
	stopFile, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatalf("could not read stop file: %v", err)
	}
	defer func() { _ = stopFile.Close() }()
	routedVehicles, err := definition.Load(scenarioFile, stopFile)
	if err != nil {
		fmt.Printf("Could not read input file %s: %v", os.Args[1], err)
		os.Exit(1)
	}

	channels := make([]<-chan routing.VehicleLocation, 0, len(routedVehicles))
	quitChannels := make([]chan int, 0, len(routedVehicles))
	for _, routedVehicle := range routedVehicles[0:] {
		routedVehicle := routedVehicle
		channel := make(chan routing.VehicleLocation)
		quitChannel := make(chan int)
		channels = append(channels, channel)
		quitChannels = append(quitChannels, quitChannel)
		go routedVehicle.StartJourney(channel, quitChannel)
	}
	consumer := channels2.Merge(channels)

	webinterface := server.NewWebInterface()
	http.HandleFunc("/sockets", webinterface.GetWebSocketHandler())

	http.Handle("/", http.FileServer(http.Dir("../webfrontend/dist/webfrontend")))

	go func() {
		for location := range consumer {
			webinterface.BroadcastJson(location)
		}
	}()

	http.ListenAndServe(":8000", nil)
}
