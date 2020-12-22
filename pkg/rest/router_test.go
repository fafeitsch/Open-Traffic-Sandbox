package rest

import (
	"encoding/json"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/bus"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockPublisher(position model.BusPosition) {
	// do nothing because it's not necessary for REST Api test.
}

func gps(coordinate ...model.Coordinate) ([]model.Coordinate, float64, error) {
	return coordinate, 250, nil
}

func TestNewRouter(t *testing.T) {
	mdl, _ := model.Init("../model/testdata/wuerzburg(fictional)")
	config := RouterConfig{
		LineModel:  mdl,
		BusModel:   mdl,
		Dispatcher: bus.NewDispatcher(mdl, mockPublisher, gps),
		Gps:        gps,
	}
	go config.Dispatcher.Run(mdl.Start())
	router := NewRouter(config)
	server := httptest.NewServer(router)
	defer server.Close()
	t.Run("get lines", func(t *testing.T) {
		resp, err := http.Get(server.URL + apiPrefix + "/lines")
		require.NoError(t, err)
		checkHeadersAndStatus(t, resp, http.StatusOK)
		var lines []restLine
		err = json.NewDecoder(resp.Body).Decode(&lines)
		require.NoError(t, err)
		assert.Equal(t, 6, len(lines), "number of lines")
		assert.Equal(t, "Zellerau - Busbahnhof", lines[3].Name, "line name")
		assert.Equal(t, "#6B8E23", lines[2].Color, "line color")
		assert.Equal(t, model.LineId("A-inbound"), lines[1].Id, "line key")
	})
	t.Run("get line", func(t *testing.T) {
		resp, err := http.Get(server.URL + apiPrefix + "/lines/A-inbound")
		require.NoError(t, err)
		checkHeadersAndStatus(t, resp, http.StatusOK)
		var line restLine
		err = json.NewDecoder(resp.Body).Decode(&line)
		require.NoError(t, err)
		assert.Equal(t, "Sanderau - Residenz - Busbahnhof", line.Name, "line name")
		assert.Equal(t, "#801818", line.Color, "line color")
		assert.Equal(t, model.LineId("A-inbound"), line.Id, "line key")
	})
	t.Run("get line route", func(t *testing.T) {
		resp, err := http.Get(server.URL + apiPrefix + "/lines/A-inbound/route")
		require.NoError(t, err)
		checkHeadersAndStatus(t, resp, http.StatusOK)
		var route [][]float64
		err = json.NewDecoder(resp.Body).Decode(&route)
		require.NoError(t, err)
		assert.Equal(t, 11, len(route), "length of the route")
		assert.Equal(t, []float64{49.7928729, 9.9364849}, route[7], "some coordinate of the route")
	})
	t.Run("get line route 404", func(t *testing.T) {
		resp, err := http.Get(server.URL + apiPrefix + "/lines/does_not_exist/route")
		require.NoError(t, err)
		checkHeadersAndStatus(t, resp, http.StatusNotFound)
		var errObj restError
		err = json.NewDecoder(resp.Body).Decode(&errObj)
		require.NoError(t, err)
		assert.NotEmpty(t, errObj.Error, "error message")
	})
	t.Run("get bus info", func(t *testing.T) {
		resp, err := http.Get(server.URL + apiPrefix + "/buses/V1/info")
		require.NoError(t, err)
		checkHeadersAndStatus(t, resp, http.StatusOK)
		var info busInfo
		err = json.NewDecoder(resp.Body).Decode(&info)
		require.NoError(t, err)
		assert.Equal(t, model.BusId("V1"), info.Id)
		assert.Equal(t, "Busbahnhof - Residenz - Sanderau", info.Line.Name, "line name")
		assert.Equal(t, "#801818", info.Line.Color, "line color")
		assert.Equal(t, model.LineId("A-outbound"), info.Line.Id, "line id")
		assert.Equal(t, "Busbahnhof - Residenz - Sanderau", info.Assignment)
	})
	t.Run("get bus info custom", func(t *testing.T) {
		resp, err := http.Get(server.URL + apiPrefix + "/buses/V2/info")
		require.NoError(t, err)
		checkHeadersAndStatus(t, resp, http.StatusOK)
		var info busInfo
		err = json.NewDecoder(resp.Body).Decode(&info)
		require.NoError(t, err)
		assert.Equal(t, model.BusId("V2"), info.Id)
		assert.Nil(t, info.Line)
		assert.Equal(t, "custom waypoint assignment", info.Assignment)
	})
	t.Run("bus route", func(t *testing.T) {
		resp, err := http.Get(server.URL + apiPrefix + "/buses/V1/route")
		require.NoError(t, err)
		checkHeadersAndStatus(t, resp, http.StatusOK)
		var route [][]float64
		err = json.NewDecoder(resp.Body).Decode(&route)
		require.NoError(t, err)
		assert.Equal(t, 11, len(route), "length of the route")
		assert.Equal(t, []float64{49.7815846, 9.9356804}, route[7], "some coordinate of the route")
	})
}

func checkHeadersAndStatus(t *testing.T, r *http.Response, status int) {
	assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type header")
	assert.Equal(t, "application/json", r.Header.Get("Accept"), "Accept header")
	assert.Equal(t, "*", r.Header.Get("Access-Control-Allow-Origin"), "cors header Access-Control-Allow-Origin")
	assert.Equal(t, "*", r.Header.Get("Access-Control-Allow-Methods"), "cors header Access-Control-Allow-Methods")
	require.Equal(t, status, r.StatusCode, "status code")
}
