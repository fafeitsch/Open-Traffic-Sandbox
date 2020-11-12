// Package model represents the static Model of a scenario and provides the Init function to
// load such a scenario.
package model

import (
	"encoding/csv"
	"fmt"
	"github.com/goccy/go-yaml"
	geojson "github.com/paulmach/go.geojson"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// BusModel is a model designed for all bus stuff.
type BusModel interface {
	Buses() []Bus
}

// Model represent the static data of a scenario. The model does not change over time, i.e. bus positions etc. are
// not stored in the model.
type Model interface {
	BusModel
	Start() Time
}

// Init loads the scenario from the provided directory and parses it.
func Init(directory string) (Model, error) {
	path := filepath.Join(directory, "scenario.yaml")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open scenario file: %v", err)
	}
	scenario := scenario{}
	err = yaml.NewDecoder(file).Decode(&scenario)
	if err != nil {
		return nil, fmt.Errorf("could not parse scenario file \"%s\": %v", path, err)
	}
	model := model{}
	model.start, err = ParseTime(scenario.Start)
	if err != nil {
		return nil, fmt.Errorf("could not parse start time \"%s\"", model.start)
	}
	stops, err := loadStops(filepath.Join(directory, scenario.StopDefinition))
	if err != nil {
		return nil, fmt.Errorf("loading the Stops from the referenced file \"%s\" failed: %v", scenario.StopDefinition, err)
	}
	model.stops = stops
	model.lines, err = loadLines(scenario, directory, stops)
	if err != nil {
		return nil, fmt.Errorf("could not load lines: %v", err)
	}
	model.buses, err = loadBuses(scenario, model.lines)
	if err != nil {
		return nil, fmt.Errorf("could not load lines: %v", err)
	}
	return &model, err
}

func loadStops(path string) (map[StopId]Stop, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	collection, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		return nil, err
	}
	stops := make(map[StopId]Stop)
	for _, feature := range collection.Features {
		id := StopId(fmt.Sprintf("%v", feature.ID))
		stop := Stop{id: id, latitude: feature.Geometry.Point[1], longitude: feature.Geometry.Point[0]}
		if name, ok := feature.Properties["name"]; ok {
			stop.name = fmt.Sprintf("%v", name)
		}
		stops[id] = stop
	}
	return stops, nil
}

func loadLines(scenario scenario, directory string, stops map[StopId]Stop) ([]Line, error) {
	loadingErrors := make([]string, 0, 0)
	result := make([]Line, 0, len(scenario.Lines))
next:
	for _, line := range scenario.Lines {
		file, err := os.Open(filepath.Join(directory, line.File))
		if err != nil {
			loadingErrors = append(loadingErrors, fmt.Sprintf("loading file line \"%s\" failed: %v", line.Id, err))
		}
		reader := csv.NewReader(file)
		reader.ReuseRecord = true
		reader.LazyQuotes = true
		stopList := make([]*Stop, 0, 0)
		departureMap := make(map[StopId][]Time)
		for data, err := reader.Read(); err == nil; data, err = reader.Read() {
			stopId := StopId(data[1])
			stop, ok := stops[stopId]
			if !ok {
				loadingErrors = append(loadingErrors, fmt.Sprintf("could not find stop \"%s\" of line \"%s\"", stopId, line.Id))
				file.Close()
				continue next
			}
			stopList = append(stopList, &stop)
			departures, err := createDepartures(data)
			if err != nil {
				loadingErrors = append(loadingErrors, fmt.Sprintf("could not parse departures of line \"%s\": %v", line.Id, err))
				continue next
			}
			departureMap[stopId] = departures
		}
		result = append(result, Line{Id: LineId(line.Id), Name: line.Name, Stops: stopList, departures: departureMap})
		file.Close()
	}
	if len(loadingErrors) != 0 {
		return nil, fmt.Errorf("%s", strings.Join(loadingErrors, ","))
	}
	return result, nil
}

func loadBuses(scenario scenario, lines []Line) ([]Bus, error) {
	lineMap := make(map[LineId]*Line)
	for _, line := range lines {
		lineMap[line.Id] = &line
	}
	result := make([]Bus, 0, len(scenario.Buses))
	for _, scenBus := range scenario.Buses {
		bus := Bus{Id: BusId(scenBus.Id)}
		assignments := make([]Assignment, 0, len(scenBus.Assignments))
		for _, asmgt := range scenBus.Assignments {
			start, err := ParseTime(asmgt.Start)
			assignment := Assignment{Departure: start}
			if err != nil {
				return nil, fmt.Errorf("could not parse time \"%s\" of bus \"%s\": %v", asmgt.Start, scenBus.Id, err)
			}
			if asmgt.Line != "" {
				line, ok := lineMap[LineId(asmgt.Line)]
				if !ok {
					return nil, fmt.Errorf("line \"%s\" of bus \"%s\" not found", asmgt.Line, scenBus.Id)
				}
				assignment.Line = line
				assignment.Name = line.Name
				waypoints := make([]WayPoint, 0, len(line.Stops))
				departures := line.TourTimes(assignment.Departure)
				if departures == nil {
					return nil, fmt.Errorf("line assignment \"%s\" of bus \"%s\" with start time \"%s\" has no equivalent in time table", line.Id, scenBus.Id, asmgt.Start)
				}
				index := 0
				for _, wp := range line.Stops {
					waypoint := WayPoint{
						IsStop:    true,
						Name:      wp.name,
						Latitude:  wp.latitude,
						Longitude: wp.longitude,
						Departure: departures[index],
					}
					waypoints = append(waypoints, waypoint)
					index = index + 1
				}
				assignment.WayPoints = waypoints
			} else {
				waypoints := make([]WayPoint, 0, len(asmgt.Coordinates))
				for _, coordinate := range asmgt.Coordinates {
					waypoint := WayPoint{
						IsStop:    false,
						Name:      "custom waypoint",
						Latitude:  coordinate[0],
						Longitude: coordinate[1],
					}
					waypoints = append(waypoints, waypoint)
				}
				assignment.WayPoints = waypoints
			}
			assignments = append(assignments, assignment)
			bus.Assignments = assignments
		}
		result = append(result, bus)
	}
	return result, nil
}

type scenario struct {
	Start          string
	StopDefinition string `json:"stopDefinition"`
	Lines          []struct {
		Name string
		Id   string
		File string
	}
	Buses []struct {
		Id          string
		Assignments []struct {
			Start       string
			Line        string
			Coordinates [][2]float64
		}
	}
}

type model struct {
	start Time
	stops map[StopId]Stop
	lines []Line
	buses []Bus
}

func (m *model) Buses() []Bus {
	return m.buses
}

func (m *model) String() string {
	result := fmt.Sprintf("Run Time: %v\n", m.start)
	result = result + fmt.Sprintf("Stops: %d\n", len(m.stops))
	result = result + fmt.Sprintf("Lines: %d\n", len(m.lines))
	result = result + fmt.Sprintf("Buses: %d", len(m.Buses()))
	return result
}

func (m *model) Start() Time {
	return m.start
}
