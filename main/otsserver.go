package main

import (
	"fmt"
	"github.com/Karmadon/gosrm"
	"github.com/fafeitsch/Open_Traffic_Sandbox/routing"
	"github.com/fafeitsch/Open_Traffic_Sandbox/server"
	geo "github.com/paulmach/go.geo"
	"github.com/twpayne/go-polyline"
	"net/http"
	"net/url"
	"os"
)

type settings struct {
	infinite bool
}

type vehicle struct {
	coordinates []routing.Coordinate
}

type simpleInput struct {
	settings settings
	vehicles []vehicle
}

type RouteService struct {
	client *gosrm.OsrmClient
}

func newRouteService() RouteService {
	localUrl := url.URL{Host: "http://localhost:5000/"}
	options := &gosrm.Options{
		Url:            localUrl,
		Service:        gosrm.ServiceRoute,
		Version:        gosrm.VersionFirst,
		Profile:        gosrm.ProfileDrivig,
		RequestTimeout: 5,
	}
	client := gosrm.NewClient(options)
	return RouteService{client: client}
}

func (r *RouteService) queryRoute(waypoints []routing.Coordinate) ([]routing.Coordinate, float64, error) {
	pointset := geo.NewPointSet()
	for index, waypoints := range waypoints {
		point := geo.NewPoint(waypoints.Lon, waypoints.Lat)
		pointset.InsertAt(index, point)
	}
	overview := "full"
	routeRequest := &gosrm.RouteRequest{
		Coordinates: *pointset,
		Overview:    &overview,
	}
	response, err := r.client.Route(routeRequest)
	if err != nil {
		return nil, 0, fmt.Errorf("Request failed: %v", err)
	}
	route := response.Routes[0]
	buffer := []byte(route.Geometry)
	coords, _, err := polyline.DecodeCoords(buffer)
	if err != nil {
		return nil, 0, fmt.Errorf("Could not decode polyline geometry: %v", err)
	}
	result := routing.PointsToCoordinates(coords)
	return result, route.Distance, nil
}

func main() {

	jsonFile, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("Could not read input file %s: %v", os.Args[1], err)
		os.Exit(1)
	}
	defer func() {
		_ = jsonFile.Close()
	}()
	routeService := newRouteService()
	routedVehicles, err := routing.LoadRoutedVehicles(jsonFile, routeService.queryRoute)

	if err != nil {
		fmt.Printf("Could not query routes for vehicles: %v", err)
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
	consumer := mergeChannels(channels)

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

func mergeChannels(channels []<-chan routing.VehicleLocation) <-chan routing.VehicleLocation {
	if len(channels) == 0 {
		return nil
	}
	if len(channels) == 1 {
		return channels[0]
	}
	m := len(channels) / 2
	channel1 := mergeChannels(channels[:m])
	channel2 := mergeChannels(channels[m:])
	return merge(channel1, channel2)
}

func merge(a, b <-chan routing.VehicleLocation) <-chan routing.VehicleLocation {
	c := make(chan routing.VehicleLocation)
	go func() {
		defer close(c)
		for a != nil || b != nil {
			select {
			case v, ok := <-a:
				if !ok {
					a = nil
					continue
				}
				c <- v
			case v, ok := <-b:
				if !ok {
					b = nil
					continue
				}
				c <- v
			}
		}
	}()
	return c
}
