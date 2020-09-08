package main

import (
	"fmt"
	channels2 "github.com/fafeitsch/Open-Traffic-Sandbox/channels"
	"github.com/fafeitsch/Open-Traffic-Sandbox/domain"
	"github.com/fafeitsch/Open-Traffic-Sandbox/osrmclient"
	"github.com/fafeitsch/Open-Traffic-Sandbox/server"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	vehicles, err := load(os.Args)
	if err != nil {
		log.Fatalf("cannot read scenario data: %v", err)
	}
	channels := make([]<-chan domain.VehicleLocation, 0, len(vehicles))
	for _, routedVehicle := range vehicles[0:] {
		ticker := time.NewTicker(100 * time.Millisecond)
		routedVehicle := routedVehicle
		routedVehicle.HeartBeat = ticker.C
		channel := make(chan domain.VehicleLocation)
		channels = append(channels, channel)
		go routedVehicle.StartJourney(channel)
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

func load(args []string) ([]domain.Vehicle, error) {
	if len(args) != 3 {
		log.Fatalf("missing scenario definition and stop definition file")
	}
	scenarioFile, err := os.Open(args[1])
	if err != nil {
		return nil, fmt.Errorf("could not read input file: %v", err)
	}
	defer func() { _ = scenarioFile.Close() }()
	stopFile, err := os.Open(args[2])
	if err != nil {
		return nil, fmt.Errorf("could not read stop file: %v", err)
	}
	defer func() { _ = stopFile.Close() }()
	stops, err := domain.LoadStops(stopFile)
	if err != nil {
		return nil, fmt.Errorf("could not parse stop file: %v", err)
	}
	vehicles, err := stops.SetupVehicles(osrmclient.NewRouteService(), scenarioFile)
	if err != nil {
		return nil, fmt.Errorf("Could not read input file %s: %v", os.Args[1], err)
	}
	return vehicles, nil
}
