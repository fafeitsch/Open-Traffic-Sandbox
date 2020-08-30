package definition

import (
	"fmt"
	"github.com/fafeitsch/Open-Traffic-Sandbox/routing"
	geojson "github.com/paulmach/go.geojson"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

type Scenario struct {
	Lines    []Line
	Start    time.Time
	Vehicles []Vehicle
}

type Line struct {
	Id    string
	Name  string
	Stops []LineStop
}

type LineStop struct {
	Arrival   int
	Departure int
	StopId    string `yaml:"stopId"`
}

type Vehicle struct {
	Id          string
	Assignments []Assignment
}

type Assignment struct {
	Start     time.Time
	End       time.Time
	Line      *string
	StartFrom *routing.Coordinate
	GoTo      *routing.Coordinate
}

func Load(scenarioReader io.Reader, stopReader io.Reader) ([]routing.RoutedVehicle, error) {
	scenario := Scenario{}
	data, err := ioutil.ReadAll(scenarioReader)
	if err != nil {
		return nil, fmt.Errorf("could not read from scenarioReader: %v", err)
	}
	err = yaml.Unmarshal(data, &scenario)
	if err != nil {
		return nil, fmt.Errorf("could not load scenario file: %v", err)
	}

	data, err = ioutil.ReadAll(stopReader)
	if err != nil {
		return nil, fmt.Errorf("could not read stop definition: %v", err)
	}
	collection, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		return nil, fmt.Errorf("could not parse stop definition: %v", err)
	}
	stops := make(map[string]routing.Coordinate)
	for _, feature := range collection.Features {
		id := fmt.Sprintf("%v", feature.ID)
		stops[id] = routing.Coordinate{Lon: feature.Geometry.Point[0], Lat: feature.Geometry.Point[1]}
	}
	lines, err := computeLines(scenario.Lines, stops)
	if err != nil {
		return nil, fmt.Errorf("could not compute lines: %v", err)
	}

	result := make([]routing.RoutedVehicle, 0, len(scenario.Vehicles))
	for _, vehicle := range scenario.Vehicles {
		if len(vehicle.Assignments) != 1 || vehicle.Assignments[0].Line == nil {
			continue
		}
		line, ok := lines[*vehicle.Assignments[0].Line]
		if !ok {
			return nil, fmt.Errorf("vehicle contains unkown line \"%s\"", *vehicle.Assignments[0].Line)
		}
		result = append(result, newRoutedVehicle(line, vehicle.Id))
	}

	return result, nil
}

func computeLines(lines []Line, stops map[string]routing.Coordinate) (map[string][]routing.Coordinate, error) {
	service := routing.NewRouteService()
	result := make(map[string][]routing.Coordinate)
	unknownStops := make([]string, 0, 0)
	for _, line := range lines {
		stopCoordinates := make([]routing.Coordinate, 0, len(line.Stops))
		for _, stop := range line.Stops {
			coordinates, ok := stops[stop.StopId]
			if !ok {
				unknownStops = append(unknownStops, fmt.Sprintf("%s (%s)", stop.StopId, line.Id))
			} else {
				stopCoordinates = append(stopCoordinates, coordinates)
			}
		}
		waypoints, _, err := service.QueryRoute(stopCoordinates)
		if err != nil {
			return nil, fmt.Errorf("could not find waypoints for line \"%s\": %v", line.Id, err)
		}
		result[line.Id] = waypoints
	}
	if len(unknownStops) != 0 {
		return nil, fmt.Errorf("could not identify the following stops: %v", strings.Join(unknownStops, ", "))
	}
	return result, nil
}

func newRoutedVehicle(geometry []routing.Coordinate, id string) routing.RoutedVehicle {
	cumulatedDistance := 0.0
	max := 0.0
	firstChainedCoordinate := routing.ChainedCoordinate{Coordinate: geometry[0]}
	coordinate := &firstChainedCoordinate
	for _, c := range geometry[1:] {
		nextChainedCoordinate := routing.ChainedCoordinate{Coordinate: c}
		coordinate.DistanceToNext = coordinate.DistanceTo(&c)
		cumulatedDistance = cumulatedDistance + coordinate.DistanceToNext
		coordinate.Next = &nextChainedCoordinate
		if coordinate.DistanceToNext > max {
			max = coordinate.DistanceToNext
		}
		coordinate = coordinate.Next
	}
	return routing.RoutedVehicle{Waypoints: &firstChainedCoordinate, Id: id, SpeedKmh: 50}
}
