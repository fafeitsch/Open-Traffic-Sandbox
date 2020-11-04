package main

import (
	"fmt"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/domain"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/osrmclient"
	"log"
	"os"
	"time"
)

func main() {
	fmt.Printf("Loading scenario file …\n")
	model, err := model.Init("samples/wuerzburg(fictive)")
	if err != nil {
		log.Fatalf("could not understand scenario file: %v", err)
	}
	fmt.Printf("Scenario loaded successfully, here is some information about it:\n")
	fmt.Println()
	fmt.Printf("%v", model)
	// scenario, err := load(os.Args)
	// if err != nil {
	// 	log.Fatalf("cannot read scenario data: %v", err)
	// }
	// channels := make([]<-chan domain.VehicleLocation, 0, len(scenario.Vehicles))
	// for _, routedVehicle := range scenario.Vehicles {
	// 	vehicle := routedVehicle
	// 	vehicle.HeartBeat = createShiftedTimer(scenario.Start)
	// 	channel := make(chan domain.VehicleLocation)
	// 	channels = append(channels, channel)
	// 	go vehicle.StartJourney(channel)
	// }
	// consumer := channels2.Merge(channels)
	//
	// clientContainer := server.NewClientContainer()
	// http.Handle("/sockets", clientContainer)
	//
	// http.Handle("/", http.FileServer(http.Dir("webfrontend/dist/webfrontend")))
	//
	// go func() {
	// 	for location := range consumer {
	// 		clientContainer.BroadcastJson(location)
	// 	}
	// }()
	// defer func() { _ = clientContainer.Close() }()
	//
	// http.ListenAndServe(":8000", nil)
}

func load(args []string) (*domain.LoadedScenario, error) {
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
	loader := domain.VehicleLoader{
		RouteService:      osrmclient.NewRouteService("http://localhost:5000/").QueryRoute,
		ExternalLocations: stops,
	}

	scenario, err := loader.SetupVehicles(scenarioFile)
	if err != nil {
		return nil, fmt.Errorf("Could not read input file %s: %v", os.Args[1], err)
	}
	return scenario, nil
}

func createShiftedTimer(start time.Time) chan time.Time {
	ticker := time.NewTicker(100 * time.Millisecond)
	difference := time.Now().Sub(start)
	result := make(chan time.Time)
	go func() {
		for tick := range ticker.C {
			result <- tick.Add(-difference)
		}
	}()
	return result
}
