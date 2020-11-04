package model

import (
	"fmt"
	routing "github.com/fafeitsch/simple-timetable-routing"
	"github.com/goccy/go-yaml"
	geojson "github.com/paulmach/go.geojson"
	"io/ioutil"
	"os"
	"path/filepath"
)

type BusModel interface {
	Buses() []Bus
}

type Model interface {
	BusModel
}

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
		return nil, fmt.Errorf("loading the stops from the referenced file \"%s\" failed: %v", scenario.StopDefinition, err)
	}
	model.stops = stops
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
		stop := Stop{id: id, latitude: feature.Geometry.Point[0], longitude: feature.Geometry.Point[1]}
		if name, ok := feature.Properties["name"]; ok {
			stop.name = fmt.Sprintf("%v", name)
		}
		stops[id] = stop
	}
	return stops, nil
}

type scenario struct {
	Start          string
	StopDefinition string `json:"stopDefinition"`
	lines          []struct {
		Name string
		Id   string
		File string
	}
	Buses []struct {
		Id          string
		Assignments []ioAssignment
	}
}

type location struct {
	Lat       float64
	Lon       float64
	Reference string
}

type ioAssignment struct {
	Start routing.Time
	Line  *string
	GoTo  *location `yaml:"goTo"`
}

type model struct {
	start Time
	stops map[StopId]Stop
}

func (m *model) Buses() []Bus {
	return []Bus{}
}

func (m *model) String() string {
	result := fmt.Sprintf("Start Time: %v\n", m.start)
	result = result + fmt.Sprintf("Stops: %d\n", len(m.stops))
	return result
}
