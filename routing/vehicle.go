package routing

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func (c *Coordinate) DistanceTo(other *Coordinate) float64 {
	earthRadius := 6371000.0 // meters
	delta1 := toRadians(c.Lat)
	delta2 := toRadians(other.Lat)
	deltaPhi := toRadians(other.Lat - c.Lat)
	deltaLambda := toRadians(other.Lon - c.Lon)
	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(delta1)*math.Cos(delta2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	atan := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * atan
}

func (c *Coordinate) toLatLngArray() [2]float64 {
	return [2]float64{c.Lat, c.Lon}
}

func newCoordinate(latLngArray [2]float64) *Coordinate {
	return &Coordinate{latLngArray[0], latLngArray[1]}
}

func toRadians(degree float64) float64 {
	return degree * (math.Pi / 180)
}

func (c *Coordinate) String() string {
	return fmt.Sprintf("[%f, %f]", c.Lat, c.Lon)
}

func PointsToCoordinates(coords [][]float64) []Coordinate {
	result := make([]Coordinate, 0, len(coords))
	for _, coord := range coords {
		result = append(result, Coordinate{Lat: coord[0], Lon: coord[1]})
	}
	return result
}

type ChainedCoordinate struct {
	Coordinate
	Next           *ChainedCoordinate
	DistanceToNext float64
}

func (c *ChainedCoordinate) ToPolyline() string {
	coordinates := make([]string, 0)
	current := c
	coordinates = append(coordinates, current.Coordinate.String())
	for current.Next != nil {
		current = current.Next
		coordinates = append(coordinates, current.Coordinate.String())
	}
	result := strings.Join(coordinates, ",")
	return "[" + result + "]"
}

func (c *ChainedCoordinate) String() string {
	return fmt.Sprintf("[%v, distanceToNext: %f, next: %v]", c.Coordinate.String(), c.DistanceToNext, c.Next.Coordinate.String())
}

type VehicleLocation struct {
	Location  [2]float64 `json:"loc"`
	VehicleId string     `json:"id"`
}

type RoutedVehicle struct {
	Waypoints *ChainedCoordinate
	SpeedKmh  float64
	Id        string
	HeartBeat <-chan time.Time
}

func (v *RoutedVehicle) StartJourney(consumer chan<- VehicleLocation) {
	speedMS := v.SpeedKmh / 3.6
	driveResult := createEmptyResult(v.Waypoints)
	last := <-v.HeartBeat
	consumer <- VehicleLocation{Location: [2]float64{v.Waypoints.Lat, v.Waypoints.Lon}, VehicleId: v.Id}
	for {
		select {
		case now, ok := <-v.HeartBeat:
			if !ok {
				close(consumer)
				return
			}
			deltaTime := now.Sub(last).Seconds()
			driven := speedMS * deltaTime
			driveResult = v.drive(driveResult.lastWp, driveResult.distanceBetween, driven)
			last = now
			location := VehicleLocation{Location: [2]float64{driveResult.location.Lat, driveResult.location.Lon}, VehicleId: v.Id}
			consumer <- location
			if driveResult.destinationReached {
				close(consumer)
				return
			}
		}
	}
}

type driveResult struct {
	location           *Coordinate
	lastWp             *ChainedCoordinate
	distanceBetween    float64
	destinationReached bool
}

func createEmptyResult(first *ChainedCoordinate) driveResult {
	return driveResult{location: &first.Coordinate, lastWp: first, distanceBetween: 0, destinationReached: false}
}

func (v *RoutedVehicle) drive(lastWp *ChainedCoordinate, distanceBetween float64, distanceToDrive float64) driveResult {
	currentDistance := distanceToDrive
	wp := lastWp
	distanceFromLast := distanceBetween
	distanceToNext := lastWp.DistanceToNext - distanceFromLast
	for currentDistance >= distanceToNext && wp.Next != nil {
		wp = wp.Next
		currentDistance = currentDistance - distanceToNext
		distanceToNext = wp.DistanceToNext
		distanceFromLast = 0.0
	}
	if wp.Next == nil {
		return driveResult{location: &Coordinate{Lat: wp.Lat, Lon: wp.Lon}, lastWp: wp, distanceBetween: 0, destinationReached: true}
	}
	lambda := (distanceFromLast + currentDistance) / wp.DistanceToNext
	deltaX := wp.Next.Lat - wp.Lat
	deltaY := wp.Next.Lon - wp.Lon
	lat := wp.Lat + lambda*deltaX
	lon := wp.Lon + lambda*deltaY
	return driveResult{location: &Coordinate{Lat: lat, Lon: lon}, lastWp: wp, distanceBetween: distanceFromLast + currentDistance}
}
