package tile

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewProxy(t *testing.T) {
	t.Run("no redirect", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(writer, "I am an osm tile (path: %s)", request.URL.Path)
		}))
		tileServer, _ := url.Parse(server.URL + "/osm-api/tiles/{z}/{x}/${y}.png")
		defer server.Close()
		proxy := NewProxy(tileServer, false)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/tile/8/652/332", nil)
		proxy.ServeHTTP(recorder, request)
		assert.Equal(t, "I am an osm tile (path: /osm-api/tiles/8/652/332.png)", recorder.Body.String())
	})
	t.Run("redirect", func(t *testing.T) {
		tileServer, _ := url.Parse("https://example.com/osm-api/tiles/{z}/{x}/${y}.png")
		proxy := NewProxy(tileServer, true)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/tile/8/652/332", nil)
		proxy.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusMovedPermanently, recorder.Code, "status code wrong")
		assert.Equal(t, "https://example.com/osm-api/tiles/8/652/332.png", recorder.Header().Get("Location"), "location header wrong")
	})
}
