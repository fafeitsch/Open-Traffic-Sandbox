package routing

import (
	"fmt"
)

type vehicle struct {
	Id          string
	Coordinates []Coordinate
}

func (v *vehicle) toRoutedVehicle(routeService func([]Coordinate) ([]Coordinate, float64, error)) (RoutedVehicle, error) {
	geometry, _, err := routeService(v.Coordinates)
	if err != nil {
		return RoutedVehicle{}, fmt.Errorf("could not get route for routing: %v", err)
	}
	cumulatedDistance := 0.0
	max := 0.0
	firstChainedCoordinate := ChainedCoordinate{Coordinate: geometry[0]}
	coordinate := &firstChainedCoordinate
	for _, c := range geometry[1:] {
		nextChainedCoordinate := ChainedCoordinate{Coordinate: c}
		coordinate.DistanceToNext = coordinate.DistanceTo(&c)
		cumulatedDistance = cumulatedDistance + coordinate.DistanceToNext
		coordinate.Next = &nextChainedCoordinate
		if coordinate.DistanceToNext > max {
			max = coordinate.DistanceToNext
		}
		coordinate = coordinate.Next
	}
	return RoutedVehicle{Waypoints: &firstChainedCoordinate, Id: v.Id, SpeedKmh: 50}, nil
}
