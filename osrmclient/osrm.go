package osrmclient

import (
	"fmt"
	"github.com/fafeitsch/Open-Traffic-Sandbox/domain"
	"github.com/karmadon/gosrm"
	geo "github.com/paulmach/go.geo"
	"github.com/twpayne/go-polyline"
	"net/url"
)

type RouteService struct {
	client *gosrm.OsrmClient
}

func NewRouteService() *RouteService {
	localUrl := url.URL{Host: "http://localhost:5000/"}
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

func (r *RouteService) QueryRoute(waypoints []domain.Coordinate) ([]domain.Coordinate, float64, error) {
	pointset := geo.NewPointSet()
	for index, waypoints := range waypoints {
		point := geo.NewPoint(waypoints.Lon, waypoints.Lat)
		pointset.InsertAt(index, point)
	}
	overview := "full"
	routeRequest := &gosrm.RouteRequest{
		Coordinates: *pointset,
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
