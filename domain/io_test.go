package domain

import (
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
	t.Run("unkown stops", func(t *testing.T) {
		stops := make(Stops)
		file, err := os.Open("testdata/testcase.yaml")
		require.NoError(t, err, "no error expexted")
		defer func() { _ = file.Close() }()
		vehicles, err := stops.SetupVehicles(service, file)
		assert.EqualError(t, err, "could not compute lines: could not identify the following stops: node/5555141689 (12-outbound), node/6805293820 (12-outbound), node/312062072 (12-outbound), node/178714408 (12-outbound), node/542261911 (12-outbound)", "error message wrong")
		assert.Nil(t, vehicles, "in case of an error the vehicles should be nil")
	})
}
