package vehicle

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Clock interface {
	Now() time.Time
	Sleep(d time.Duration)
}

type SystemClock struct {
}

func (s *SystemClock) Now() time.Time {
	return time.Now()
}

func (s *SystemClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

type Coordinate struct {
	Lat float64
	Lon float64
}

func (c *Coordinate) DistanceTo(other *Coordinate) float64 {
	earthRadius := 6371000.0 // metres
	φ1 := toRadians(c.Lat)
	φ2 := toRadians(other.Lat)
	Δφ := toRadians(other.Lat - c.Lat)
	Δλ := toRadians(other.Lon - c.Lon)
	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Sin(Δλ/2)*math.Sin(Δλ/2)
	husten := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * husten
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
	Location  *Coordinate
	VehicleId int
}

type RoutedVehicle struct {
	Waypoints *ChainedCoordinate
	SpeedKmh  float64
	Id        int
}

func (v *RoutedVehicle) StartJourney(consumer chan<- VehicleLocation, quit <-chan int) {
	v.startJourneyWithClock(&SystemClock{}, consumer, quit)
}

func (v *RoutedVehicle) startJourneyWithClock(clock Clock, consumer chan<- VehicleLocation, quit <-chan int) {
	speedMS := v.SpeedKmh / 3.6
	driveResult := createEmptyResult(v.Waypoints)
	time.Sleep(10 * time.Second)
	last := clock.Now()
	fmt.Printf("%s\n", v.Waypoints.ToPolyline())
	consumer <- VehicleLocation{Location: driveResult.location, VehicleId: v.Id}
	for !driveResult.destinationReached {
		select {
		case <-quit:
			close(consumer)
			return
		default:
			clock.Sleep(40 * time.Millisecond)
			now := clock.Now()
			deltaTime := now.Sub(last).Seconds()
			driven := speedMS * deltaTime
			driveResult = v.drive(driveResult.lastWp, driveResult.distanceBetween, driven)
			last = now
			location := VehicleLocation{Location: driveResult.location, VehicleId: v.Id}
			consumer <- location
		}
	}
	close(consumer)
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
