package main

import (
	"fmt"
	"github.com/Karmadon/gosrm"
	"github.com/paulmach/go.geo"
	"github.com/twpayne/go-polyline"
	"net/url"
)

func main() {
	localUrl := url.URL{Host: "http://localhost:5000/"}
	options := &gosrm.Options{
		Url:            localUrl,
		Service:        gosrm.ServiceRoute,
		Version:        gosrm.VersionFirst,
		Profile:        gosrm.ProfileDrivig,
		RequestTimeout: 5,
	}

	client := gosrm.NewClient(options)

	overview := "full"
	routeRequest := &gosrm.RouteRequest{
		Coordinates: geo.PointSet{
			{9.938611, 49.792778}, {9.863611, 49.835833},
		},
		Overview: &overview,
	}

	response, err := client.Route(routeRequest)
	if err != nil {
		panic(err.Error())
	}

	buf := []byte(response.Routes[0].Geometry)
	coords, _, _ := polyline.DecodeCoords(buf)
	distance := 0.0
	for index, coordinate := range coords[1:] {
		lastCoordinate := coords[index]
		pointset := geo.PointSet{{lastCoordinate[1], lastCoordinate[0]}, {coordinate[1], coordinate[0]}}
		request := &gosrm.RouteRequest{
			Coordinates: pointset,
			Overview:    &overview,
		}
		r, e := client.Route(request)
		if e != nil {
			panic(e.Error())
		}
		curr := r.Routes[0].Distance
		distance = distance + curr
		fmt.Printf("Distance: %f\n", r.Routes[0].Distance)
	}

	fmt.Printf("Summed Distance: %f, Queried Distance: %f", distance, response.Routes[0].Distance)
}
