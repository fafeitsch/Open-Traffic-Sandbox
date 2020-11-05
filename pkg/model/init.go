package model

import (
	"encoding/csv"
	"fmt"
	routing "github.com/fafeitsch/simple-timetable-routing"
	"github.com/goccy/go-yaml"
	geojson "github.com/paulmach/go.geojson"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
	model.lines, err = loadLines(scenario, directory, stops)
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
		stop := Stop{id: id, latitude: feature.Geometry.Point[0], longitude: feature.Geometry.Point[1]}
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
		result = append(result, Line{id: LineId(line.Id), name: line.Name, stops: stopList, departures: departureMap})
		file.Close()
	}
	if len(loadingErrors) != 0 {
		return nil, fmt.Errorf("%s", strings.Join(loadingErrors, ","))
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
	lines []Line
}

func (m *model) Buses() []Bus {
	return []Bus{}
}

func (m *model) String() string {
	result := fmt.Sprintf("Start Time: %v\n", m.start)
	result = result + fmt.Sprintf("Stops: %d\n", len(m.stops))
	result = result + fmt.Sprintf("Lines: %d", len(m.lines))
	return result
}
