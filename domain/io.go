package domain

import (
	"fmt"
	geojson "github.com/paulmach/go.geojson"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

type scenario struct {
	Lines    []line
	Start    time.Time
	Vehicles []vehicle
}

type line struct {
	Id        string
	Name      string
	Stops     []lineStop
	waypoints Coordinates
}

func (l *line) toRoutingLine() *Line {
	return &Line{Id: l.Id, Name: l.Name, Waypoints: l.waypoints}
}

type lineStop struct {
	Arrival   int
	Departure int
	StopId    string `yaml:"stopId"`
}

type vehicle struct {
	Id          string
	Assignments assignments
}

type assignment struct {
	Start     time.Time
	Line      *string
	StartFrom *Coordinate
	GoTo      *Coordinate
}

type assignments []assignment

func (a assignments) toRoutingAssignments(lines map[string]line) ([]Assignment, error) {
	result := make([]Assignment, 0, len(a))
	for _, assignment := range a {
		var res Assignment
		if assignment.Line != nil {
			line, ok := lines[*assignment.Line]
			if !ok {
				return nil, fmt.Errorf("line with name %s is not defined", *assignment.Line)
			}
			res = Assignment{Line: line.toRoutingLine()}
		} else if assignment.GoTo != nil {
			res = Assignment{GoTo: assignment.GoTo}
		}
		res.Start = assignment.Start
		result = append(result, res)
	}
	return result, nil
}

// Stops is a map which stores the OSM node ids and the coordinate the respective node is located at.
type Stops map[string]Coordinate

// LoadStops reads all features from stopReader. StopReader must contain a valid GeoJSON with valid
// features. The return format is a map containing the nodes ids of all stops and their coordinates.
// This method does not filter the GeoJSON. If the GeoJSON contains non-bus-stop features, they will also be returned.
func LoadStops(stopReader io.Reader) (Stops, error) {
	data, err := ioutil.ReadAll(stopReader)
	if err != nil {
		return nil, fmt.Errorf("could not read stop definition: %v", err)
	}
	collection, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		return nil, fmt.Errorf("could not parse stop definition: %v", err)
	}
	stops := make(map[string]Coordinate)
	for _, feature := range collection.Features {
		id := fmt.Sprintf("%v", feature.ID)
		stops[id] = Coordinate{Lon: feature.Geometry.Point[0], Lat: feature.Geometry.Point[1]}
	}
	return stops, nil
}

// RouteService is an interface capable of computing detailed waypoints between the provided waypoints.
type RouteService func(Coordinates) (Coordinates, float64, error)

// SetupVehicles reads the scenario from the scenario reader and precomputes the routes the vehicles must make.
// For computing the routes, the routeService is used.
func (s Stops) SetupVehicles(routeService RouteService, scenarioReader io.Reader) ([]Vehicle, error) {
	scenario := scenario{}
	data, err := ioutil.ReadAll(scenarioReader)
	if err != nil {
		return nil, fmt.Errorf("could not read from scenarioReader: %v", err)
	}
	err = yaml.Unmarshal(data, &scenario)
	if err != nil {
		return nil, fmt.Errorf("could not load scenario file: %v", err)
	}

	lines, err := computeLines(routeService, scenario.Lines, s)
	if err != nil {
		return nil, fmt.Errorf("could not compute lines: %v", err)
	}

	result := make([]Vehicle, 0, len(scenario.Vehicles))
	for _, vehicle := range scenario.Vehicles {
		assignments, err := vehicle.Assignments.toRoutingAssignments(lines)
		if err != nil {
			return nil, fmt.Errorf("could not build assignments for vehicle \"%s\"", vehicle.Id)
		}
		created := Vehicle{Id: vehicle.Id, Assignments: assignments, SpeedKmh: 20}
		result = append(result, created)
	}
	return result, nil
}

func computeLines(service RouteService, lines []line, stops map[string]Coordinate) (map[string]line, error) {
	result := make(map[string]line)
	unknownStops := make([]string, 0, 0)
	for _, line := range lines {
		stopCoordinates := make([]Coordinate, 0, len(line.Stops))
		for _, stop := range line.Stops {
			coordinates, ok := stops[stop.StopId]
			if !ok {
				unknownStops = append(unknownStops, fmt.Sprintf("%s (%s)", stop.StopId, line.Id))
			} else {
				stopCoordinates = append(stopCoordinates, coordinates)
			}
		}
		waypoints, _, err := service(stopCoordinates)
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
