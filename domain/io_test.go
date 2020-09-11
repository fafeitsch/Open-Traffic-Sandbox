package domain

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestLoadStops(t *testing.T) {
	t.Run("invalid format", func(t *testing.T) {
		stops, err := LoadStops(strings.NewReader("not a valid json"))
		assert.Nil(t, stops, "stops should be nil in case of an error")
		assert.EqualError(t, err, "could not parse stop definition: invalid character 'o' in literal null (expecting 'u')", "err message not correct")
	})
	t.Run("success", func(t *testing.T) {
		file, err := os.Open("testdata/stops.geojson")
		require.NoError(t, err, "error should be nil")
		defer func() { _ = file.Close() }()
		stops, err := LoadStops(file)
		require.NoError(t, err, "error should be nil")
		assert.Equal(t, 690, len(stops))
		assert.Equal(t, stops["node/201065598"], Coordinate{Lat: 49.7975861, Lon: 9.9336463}, "coordinates not read correctly")
	})
}

func TestStops_SetupVehicles(t *testing.T) {
	service := func(waypoints Coordinates) (Coordinates, float64, error) {
		return waypoints, 0, nil
	}
	t.Run("unknown stops", func(t *testing.T) {
		stops := make(Stops)
		file, err := os.Open("testdata/testcase1.yaml")
		require.NoError(t, err, "no error expected")
		defer func() { _ = file.Close() }()
		vehicles, err := stops.SetupVehicles(service, file)
		assert.EqualError(t, err, "could not compute lines: could not identify the following stops: node/5555141689 (12-outbound), node/6805293820 (12-outbound), node/312062072 (12-outbound), node/178714408 (12-outbound), node/542261911 (12-outbound)", "error message wrong")
		assert.Nil(t, vehicles, "in case of an error the vehicles should be nil")
	})
	t.Run("unparsable reader", func(t *testing.T) {
		stops := make(Stops)
		vehicles, err := stops.SetupVehicles(service, strings.NewReader("{not a valid json/yaml"))
		assert.Nil(t, vehicles, "result should be nil in case of an error")
		assert.EqualError(t, err, "could not load scenario file: yaml: line 1: did not find expected ',' or '}'")
	})
	t.Run("success", func(t *testing.T) {
		stopsFile, err := os.Open("testdata/stops.geojson")
		require.NoError(t, err)
		defer func() { _ = stopsFile.Close() }()
		stops, err := LoadStops(stopsFile)
		require.NoError(t, err)
		require.NotNil(t, stops)
		file, err := os.Open("testdata/testcase1.yaml")
		require.NoError(t, err)
		defer func() { _ = file.Close() }()
		vehicles, err := stops.SetupVehicles(service, file)
		require.NoError(t, err)
		assert.Equal(t, 1, len(vehicles), "number of loaded vehicles")
		vehicle := vehicles[0]
		assert.Equal(t, 3, len(vehicle.Assignments), "number of assignments")
		assert.Nil(t, vehicle.Assignments[0].GoTo, "should be a line assignment with Goto == nil")
		line := vehicle.Assignments[0].Line
		assert.Equal(t, 5, len(line.Waypoints), "waypoints of line should be extracted correctly")
		assert.Equal(t, "12-outbound", line.Id, "id of line should be extracted correctly")
		assert.Equal(t, "Busbahnhof - Lindleinsmühle - Versbach", line.Name, "name of line should be extracted correctly")
		assert.Equal(t, &Coordinate{Lat: 13.03, Lon: 23.93}, vehicle.Assignments[1].GoTo)
		assert.Nil(t, vehicle.Assignments[1].Line, "line should be nil if goto is set")
		assert.NotNil(t, vehicle.Assignments[2].GoTo, "should be a Goto-assignment")
	})
	t.Run("unknown line", func(t *testing.T) {
		stopsFile, err := os.Open("testdata/stops.geojson")
		require.NoError(t, err)
		defer func() { _ = stopsFile.Close() }()
		stops, err := LoadStops(stopsFile)
		require.NoError(t, err)
		require.NotNil(t, stops)
		file, err := os.Open("testdata/testcase2.yaml")
		require.NoError(t, err)
		defer func() { _ = file.Close() }()
		vehicles, err := stops.SetupVehicles(service, file)
		assert.EqualError(t, err, "could not build assignments for vehicle \"V1\": line with name \"13-not found\" is not defined")
		assert.Nil(t, vehicles, "result should be nil in case of an error")
	})
	t.Run("service error", func(t *testing.T) {
		service = func(coordinates Coordinates) (Coordinates, float64, error) {
			return nil, 0, fmt.Errorf("planned error")
		}
		stopsFile, err := os.Open("testdata/stops.geojson")
		require.NoError(t, err)
		defer func() { _ = stopsFile.Close() }()
		stops, err := LoadStops(stopsFile)
		require.NoError(t, err)
		require.NotNil(t, stops)
		file, err := os.Open("testdata/testcase1.yaml")
		require.NoError(t, err)
		defer func() { _ = file.Close() }()
		vehicles, err := stops.SetupVehicles(service, file)
		assert.EqualError(t, err, "could not compute lines: could not find waypoints for line \"12-outbound\": planned error")
		assert.Nil(t, vehicles, "result should be nil in case of an error")
	})
}
