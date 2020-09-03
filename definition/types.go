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
	Id        string
	Name      string
	Stops     []LineStop
	waypoints routing.Coordinates
}

func (l *Line) toRoutingLine() *routing.Line {
	return &routing.Line{Id: l.Id, Name: l.Name, Waypoints: l.waypoints}
}

type LineStop struct {
	Arrival   int
	Departure int
	StopId    string `yaml:"stopId"`
}

type Vehicle struct {
	Id          string
	Assignments Assignments
}

type Assignment struct {
	Start     time.Time
	Line      *string
	StartFrom *routing.Coordinate
	GoTo      *routing.Coordinate
}

type Assignments []Assignment

func (a Assignments) toRoutingAssignments(lines map[string]Line) ([]routing.Assignment, error) {
	result := make([]routing.Assignment, 0, len(a))
	for _, assignment := range a {
		var res routing.Assignment
		if assignment.Line != nil {
			line, ok := lines[*assignment.Line]
			if !ok {
				return nil, fmt.Errorf("line with name %s is not defined", *assignment.Line)
			}
			res = routing.Assignment{Line: line.toRoutingLine()}
		} else if assignment.GoTo != nil {
			res = routing.Assignment{GoTo: assignment.GoTo}
		}
		res.Start = assignment.Start
		result = append(result, res)
	}
	return result, nil
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
		assignments, err := vehicle.Assignments.toRoutingAssignments(lines)
		if err != nil {
			return nil, fmt.Errorf("could not build assignments for vehicle \"%s\"", vehicle.Id)
		}
		created := routing.RoutedVehicle{Id: vehicle.Id, Assignments: assignments, SpeedKmh: 20}
		result = append(result, created)
	}

	return result, nil
}

func computeLines(lines []Line, stops map[string]routing.Coordinate) (map[string]Line, error) {
	service := routing.NewRouteService()
	result := make(map[string]Line)
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
		line.waypoints = waypoints
		result[line.Id] = line
	}
	if len(unknownStops) != 0 {
		return nil, fmt.Errorf("could not identify the following stops: %v", strings.Join(unknownStops, ", "))
	}
	return result, nil
}
