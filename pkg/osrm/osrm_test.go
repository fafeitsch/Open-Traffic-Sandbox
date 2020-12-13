package osrm

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRouteService_QueryRoute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		successOsrm := func(writer http.ResponseWriter, request *http.Request) {
			assert.Equal(t, "full", request.URL.Query()["overview"][0], "overview query param should be full")
			assert.Equal(t, "false", request.URL.Query()["generate_hints"][0], "generate hints should be disabled")
			fmt.Printf("\n%v\n", request.URL.Query())
			assert.Equal(t, "/route/v1/driving/polyline(_qo%5D_%7Brc@_seK_seK)", request.URL.Path, "path must be corrected")
			response, err := os.Open("testdata/osrmresult.json")
			require.NoError(t, err)
			_, _ = io.Copy(writer, response)
		}
		server := httptest.NewServer(http.HandlerFunc(successOsrm))
		defer server.Close()
		service := NewRouteService(server.URL + "/")
		route, length, err := service(&coordinate{5, 6}, &coordinate{7, 8})
		require.Nil(t, err)
		assert.Equal(t, 7634.4, length, "length was not computed correctly")
		assert.Equal(t, 396, len(route), "number of waypoints wrong")
	})
	t.Run("failure", func(t *testing.T) {
		successOsrm := func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write([]byte("{\n  \"code\": \"TooBig\"}"))
		}
		server := httptest.NewServer(http.HandlerFunc(successOsrm))
		defer server.Close()
		service := NewRouteService(server.URL + "/")
		route, length, err := service(coordinate{5, 6}, coordinate{7, 8})
		require.EqualError(t, err, "Request failed: [GOSRM][ERROR]: The request size violates one of the service specific request size restrictions")
		assert.Equal(t, 0.0, length, "length should be 0 in case of an error")
		assert.Nil(t, route, "route should be nil in case of an error")
	})
}
