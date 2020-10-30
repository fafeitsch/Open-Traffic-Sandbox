package osrmclient

import (
	"fmt"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/domain"
	"github.com/karmadon/gosrm"
	geo "github.com/paulmach/go.geo"
	"github.com/twpayne/go-polyline"
	"net/url"
)

// RouteService is a OSRM client providing methods to
// query data from an OSRM server.
type RouteService struct {
	client *gosrm.OsrmClient
}

// NewRouteService creates a route service and sets the most important
// settings for connecting to the OSRM server. The connection parameter
// denotes the address to connect to, i.e. http://localhost:5000/
func NewRouteService(connection string) *RouteService {
	localUrl := url.URL{Host: connection}
	options := &gosrm.Options{
		Url:            localUrl,
		Service:        gosrm.ServiceRoute,
		Version:        gosrm.VersionFirst,
		Profile:        gosrm.ProfileDriving,
		RequestTimeout: 5,
	}
	client := gosrm.NewClient(options)
	return &RouteService{client: client}
}

// QueryRoute sends a request to the OSRM server, asking for the shortest path
// to connect the given waypoints the the defined order (no, this is not a TSP here :)).
// When the request is responded successfully, the shortest path is returned as well as the length
// of the shortest path. Otherwise, a non-nil error is returned.
func (r *RouteService) QueryRoute(waypoints domain.Coordinates) (domain.Coordinates, float64, error) {
	pointSet := geo.NewPointSet()
	for index, waypoints := range waypoints {
		point := geo.NewPoint(waypoints.Lon, waypoints.Lat)
		pointSet.InsertAt(index, point)
	}
	overview := "full"
	routeRequest := &gosrm.RouteRequest{
		Coordinates: *pointSet,
		Overview:    &overview,
	}
	response, err := r.client.Route(routeRequest)
	if err != nil {
		return nil, 0, fmt.Errorf("Request failed: %v", err)
	}
	route := response.Routes[0]
	buffer := []byte(route.Geometry)
	coords, _, err := polyline.DecodeCoords(buffer)
	if err != nil {
		return nil, 0, fmt.Errorf("Could not decode polyline geometry: %v", err)
	}
	result := domain.PointsToCoordinates(coords)
	return result, route.Distance, nil
}
