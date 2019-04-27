package route

import (
	"fmt"
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

func (c *Coordinate) String() string {
	return fmt.Sprintf("[%f, %f]", c.Lat, c.Lon)
}

type ChainedCoordinate struct {
	Coordinate
	Next           *ChainedCoordinate
	DistanceToNext float64
}

type VehicleLocation struct {
	Location *Coordinate
	Vehicle  *RoutedVehicle
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
	last := clock.Now()
	speedMS := v.SpeedKmh / 3.6
	driveResult := createEmptyResult(v.Waypoints)
	for !driveResult.destinationReached {
		select {
		case <-quit:
			close(consumer)
			return
		default:
			clock.Sleep(50 * time.Millisecond)
			now := clock.Now()
			deltaTime := now.Sub(last).Seconds()
			driven := speedMS * deltaTime
			driveResult = v.drive(driveResult.lastWp, driveResult.distanceBetween, driven)
			last = now
			location := VehicleLocation{Location: driveResult.location, Vehicle: v}
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
	lambda := (distanceFromLast + currentDistance) / distanceToNext
	deltaX := wp.Next.Lat - wp.Lat
	deltaY := wp.Next.Lon - wp.Lon
	lat := wp.Lat + lambda*deltaX
	lon := wp.Lon + lambda*deltaY
	return driveResult{location: &Coordinate{Lat: lat, Lon: lon}, lastWp: wp, distanceBetween: distanceFromLast + currentDistance}
}
