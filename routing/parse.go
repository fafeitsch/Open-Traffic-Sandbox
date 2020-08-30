package routing

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

type settings struct {
	Infinite bool `json:"infinite"`
}

type vehicle struct {
	Id          string       `json:"id"`
	Coordinates []Coordinate `json:"coordinates"`
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

type simpleConfig struct {
	Settings settings  `json:"settings"`
	Vehicles []vehicle `json:"vehicles"`
}

func parseJson(reader io.Reader) (simpleConfig, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return simpleConfig{}, fmt.Errorf("could not read file: %v", err)
	}
	var input simpleConfig
	err = json.Unmarshal(bytes, &input)
	if err != nil {
		return simpleConfig{}, fmt.Errorf("could not parse json: %v", err)
	}
	return input, err
}

func LoadRoutedVehicles(reader io.Reader, routeService func([]Coordinate) ([]Coordinate, float64, error)) ([]RoutedVehicle, error) {
	input, err := parseJson(reader)
	if err != nil {
		return nil, err
	}
	result := make([]RoutedVehicle, 0, len(input.Vehicles))
	for _, v := range input.Vehicles {
		routedVehicle, err := v.toRoutedVehicle(routeService)
		if err != nil {
			return nil, err
		}
		result = append(result, routedVehicle)
	}
	return result, nil
}
