package tile

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

var zIndex = regexp.MustCompile("\\$?{z}")
var xIndex = regexp.MustCompile("\\$?{x}")
var yIndex = regexp.MustCompile("\\$?{y}")

// NewProxy creates a new proxy for the tile server. If the redirect parameter is false, then the returned proxy
// acts as reverse proxy and actively fetches the tiles from the tileUrl. If redirect is true, then the returned handler
// just sends 301 headers with the actual location of the requested tile.
func NewProxy(tileUrl *url.URL, redirect bool) http.Handler {
	if !redirect {
		director := func(req *http.Request) {
			req.Header.Add("X-Forwarded-Host", req.Host)
			req.Header.Add("X-Origin-Host", tileUrl.Host)
			req.URL.Host = tileUrl.Host
			req.URL.Scheme = tileUrl.Scheme
			newPath := buildNewPath(req.URL.Path, tileUrl)
			req.URL.Path = newPath
		}
		proxy := &httputil.ReverseProxy{Director: director}
		return proxy
	} else {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			target := buildNewPath(request.URL.Path, tileUrl)
			writer.Header().Set("Location", tileUrl.Scheme+"://"+tileUrl.Host+target)
			writer.WriteHeader(http.StatusMovedPermanently)
		})
	}
}

func buildNewPath(original string, tileUrl *url.URL) string {
	parts := strings.Split(original, "/")
	newPath := zIndex.ReplaceAllString(tileUrl.Path, parts[2])
	newPath = xIndex.ReplaceAllString(newPath, parts[3])
	return yIndex.ReplaceAllString(newPath, parts[4])
}
