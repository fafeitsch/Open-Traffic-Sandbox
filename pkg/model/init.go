// Package model represents the static Model of a scenario and provides the Init function to
// load such a scenario.
package model

import (
	"fmt"
	"github.com/goccy/go-yaml"
	geojson "github.com/paulmach/go.geojson"
	"io/ioutil"
	"os"
	"path/filepath"
)

// BusModel is a model designed for all bus stuff.
type BusModel interface {
	Buses() []Bus
	Bus(BusId) (*Bus, bool)
}

// LineModel is model designed for line management
type LineModel interface {
	Lines() []Line
	Line(LineId) (Line, bool)
}

// Model represent the static data of a scenario. The model does not change over time, i.e. bus positions etc. are
// not stored in the model.
type Model interface {
	BusModel
	LineModel
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
	stops := make(map[StopId]Stop)
	for _, stopFile := range scenario.StopDefinitions {
		err := loadStops(filepath.Join(directory, stopFile), stops)
		if err != nil {
			return nil, fmt.Errorf("loading the Stops from the referenced file \"%s\" failed: %v", stopFile, err)
		}
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

func loadStops(path string, stops map[StopId]Stop) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	collection, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		return err
	}
	for _, feature := range collection.Features {
		id := StopId(fmt.Sprintf("%v", feature.ID))
		stop := Stop{Id: id, WayPoint: WayPoint{IsRealStop: true, Latitude: feature.Geometry.Point[1], Longitude: feature.Geometry.Point[0]}}
		if name, ok := feature.Properties["name"]; ok {
			stop.Name = fmt.Sprintf("%v", name)
		}
		stops[id] = stop
	}
	return nil
}

type scenario struct {
	Start           string
	StopDefinitions []string `json:"stopDefinitions"`
	Lines           []struct {
		Name  string
		Color string
		Id    string
		File  string
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
	lines map[LineId]Line
	buses map[BusId]Bus
}

// Buses returns a slice of all busses in this model.
func (m *model) Buses() []Bus {
	result := make([]Bus, 0, len(m.buses))
	for _, bus := range m.buses {
		result = append(result, bus)
	}
	return result
}

// Bus returns a pointer to the bus with the given id. If the bus does not exist, then the second return variable is false.
func (m *model) Bus(id BusId) (*Bus, bool) {
	bus, ok := m.buses[id]
	return &bus, ok
}

// Lines returns a slice of all lines in this model.
func (m *model) Lines() []Line {
	result := make([]Line, 0, len(m.lines))
	for _, line := range m.lines {
		result = append(result, line)
	}
	return result
}

func (m *model) Line(s LineId) (Line, bool) {
	line, ok := m.lines[s]
	return line, ok
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
