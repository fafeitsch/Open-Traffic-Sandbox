package domain

import (
	"fmt"
	"github.com/goccy/go-yaml"
	geojson "github.com/paulmach/go.geojson"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"strings"
	"time"
)

type scenario struct {
	Lines    []line
	Start    time.Time
	Vehicles []vehicle
}

type line struct {
	Id    string
	Name  string
	Stops []lineStop
	legs  []Coordinates
}

func (l *line) toAssignments() []Assignment {
	randomWaiting := func(now time.Time, vehicle *Vehicle, next *Assignment) {
		if next == nil {
			return
		}
		random := rand.New(rand.NewSource(time.Now().Unix()))
		seconds := random.NormFloat64()*10 + 20
		seconds = math.Max(0, seconds)
		seconds = math.Min(600, seconds)
		then := now.Add(time.Duration(seconds) * time.Second)
		if then.After(next.Start) {
			fmt.Printf("waiting %d seconds", int(seconds))
			if now.Before(next.Start) {
				difference := then.Sub(next.Start)
				next.Start = now.Add(difference)
			} else {
				next.Start = then
			}
		}
	}
	result := make([]Assignment, 0, len(l.legs))
	for _, leg := range l.legs {
		assignment := Assignment{Waypoints: leg, destinationHandler: randomWaiting}
		result = append(result, assignment)
	}
	return result
}

type lineStop struct {
	Arrival   int
	Departure int
	Location  location
}

type location struct {
	Lat       float64
	Lon       float64
	Reference string
}

type vehicle struct {
	Id          string
	Assignments assignments
}

type assignment struct {
	Start time.Time
	Line  *string
	GoTo  *location `yaml:"goTo"`
}

type assignments []assignment

func (a assignments) toRoutingAssignments(service RouteService, lines map[string]line) ([]Assignment, error) {
	result := make([]Assignment, 0, len(a))
	for index, assignment := range a {
		var res Assignment
		res.Start = assignment.Start
		if assignment.Line != nil {
			line, ok := lines[*assignment.Line]
			if !ok {
				return nil, fmt.Errorf("line with name \"%s\" is not defined", *assignment.Line)
			}
			for _, leg := range line.toAssignments() {
				result = append(result, leg)
			}
		} else if assignment.GoTo != nil {
			// If the GoTo assignment is not the first one, we have to find the
			// route from the last waypoint to the GoTo-coordinates …
			if index > 0 && len(result[index-1].Waypoints) > 0 {
				lastWaypoints := result[index-1].Waypoints
				lastWaypoint := lastWaypoints[len(lastWaypoints)-1]
				waypoints, _, err := service(Coordinates{lastWaypoint, {assignment.GoTo.Lat, assignment.GoTo.Lon}})
				if err != nil {
					return nil, fmt.Errorf("could not query route for GoTo-Assignment (index %d): %v", index, err)
				}
				res.Waypoints = waypoints
			} else {
				// … Otherwise, there is no previous waypoint and we simply beam the vehicle to the GoTo point.
				res.Waypoints = Coordinates{{Lat: assignment.GoTo.Lat, Lon: assignment.GoTo.Lon}}
			}
			result = append(result, res)
		}
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

func (s Stops) resolve(location location) (*Coordinate, error) {
	if location.Reference == "" {
		return &Coordinate{Lat: location.Lat, Lon: location.Lon}, nil
	}
	if stop, ok := s[location.Reference]; ok {
		return &stop, nil
	}
	return nil, fmt.Errorf("unresolvable location \"%s\"", location.Reference)
}

// RouteService is an interface capable of computing detailed waypoints between the provided waypoints.
type RouteService func(Coordinates) (Coordinates, float64, error)

type VehicleLoader struct {
	RouteService      RouteService
	ExternalLocations Stops
}

// SetupVehicles reads the scenario from the scenario reader and precomputes the routes the vehicles must make.
// For computing the routes, the routeService is used.
func (v *VehicleLoader) SetupVehicles(scenarioReader io.Reader) ([]Vehicle, error) {
	scenario := scenario{}
	data, err := ioutil.ReadAll(scenarioReader)
	if err != nil {
		return nil, fmt.Errorf("could not read from scenarioReader: %v", err)
	}
	err = yaml.Unmarshal(data, &scenario)
	if err != nil {
		return nil, fmt.Errorf("could not load scenario file: %v", err)
	}

	err = v.resolveLocations(&scenario)
	if err != nil {
		return nil, err
	}
	lines, err := v.resolveLines(scenario.Lines)
	if err != nil {
		return nil, fmt.Errorf("could not compute lines: %v", err)
	}

	result := make([]Vehicle, 0, len(scenario.Vehicles))
	for _, vehicle := range scenario.Vehicles {
		assignments, err := vehicle.Assignments.toRoutingAssignments(v.RouteService, lines)
		if err != nil {
			return nil, fmt.Errorf("could not build assignments for vehicle \"%s\": %v", vehicle.Id, err)
		}
		created := Vehicle{Id: vehicle.Id, Assignments: assignments, SpeedKmh: 50}
		result = append(result, created)
	}
	return result, nil
}

func (v *VehicleLoader) resolveLocations(scenario *scenario) error {
	errors := make([]string, 0, 0)
	for _, line := range scenario.Lines {
		for _, stop := range line.Stops {
			resolved, err := v.ExternalLocations.resolve(stop.Location)
			if err != nil {
				errors = append(errors, err.Error())
			} else {
				stop.Location.Lat = resolved.Lat
				stop.Location.Lon = resolved.Lon
			}
		}
	}
	for _, vehicle := range scenario.Vehicles {
		for _, assignment := range vehicle.Assignments {
			if assignment.GoTo != nil {
				resolved, err := v.ExternalLocations.resolve(*assignment.GoTo)
				if err != nil {
					errors = append(errors, err.Error())
				} else {
					assignment.GoTo.Lat = resolved.Lat
					assignment.GoTo.Lon = resolved.Lon
				}
			}
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("could not resolve some locations: %s", strings.Join(errors, ", "))
	}
	return nil
}

func (v *VehicleLoader) resolveLines(lines []line) (map[string]line, error) {
	result := make(map[string]line)
	for _, line := range lines {
		legs := make([]Coordinates, 0, len(line.Stops)-1)
		for index, currentStop := range line.Stops[0 : len(line.Stops)-1] {
			currentCoordinate, _ := v.ExternalLocations.resolve(currentStop.Location)
			nextCoordinate, _ := v.ExternalLocations.resolve(line.Stops[index+1].Location)
			leg, _, err := v.RouteService(Coordinates{*currentCoordinate, *nextCoordinate})
			if err != nil {
				return nil, fmt.Errorf("could not find routes for line \"%s\", %dth leg: %v", line.Id, index+1, err)
			}
			legs = append(legs, leg[0:])
		}
		line.legs = legs
		result[line.Id] = line
	}
	return result, nil
}
