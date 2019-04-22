package route

import "time"

type Coordinate struct {
	Lat float64
	Lon float64
}

type ChainedCoordinate struct {
	Coordinate
	Next           *ChainedCoordinate
	DistanceToNext float64
}

func (c *ChainedCoordinate) computeRemainingDistance() float64 {
	distance := 0.0
	for c.Next != nil {
		distance = distance + c.DistanceToNext
		c = c.Next
	}
	return distance
}

type VehicleLocation struct {
	Location Coordinate
	Vehicle  *RoutedVehicle
}

type RoutedVehicle struct {
	Waypoints *ChainedCoordinate
	SpeedKmh  float64
	Id        int
}

func (v *RoutedVehicle) StartJourney(listener chan VehicleLocation) {
	driven := 0.0
	totalDistance := v.Waypoints.computeRemainingDistance()
	last := time.Now()
	time.Sleep(50 * time.Millisecond)
	speed := v.SpeedKmh / 3.6
	_ = v.Waypoints
	_ = 0.0
	for driven < totalDistance {
		now := time.Now()
		deltaTime := now.Sub(last).Seconds()
		driven = driven + (speed * deltaTime)
	}

}

type driveResult struct {
	location        *Coordinate
	lastWp          *ChainedCoordinate
	distanceBetween float64
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
		return driveResult{location: &Coordinate{Lat: wp.Lat, Lon: wp.Lon}, lastWp: wp, distanceBetween: 0}
	}
	lambda := (distanceFromLast + currentDistance) / distanceToNext
	deltaX := wp.Next.Lat - wp.Lat
	deltaY := wp.Next.Lon - wp.Lon
	lat := wp.Lat + lambda*deltaX
	lon := wp.Lon + lambda*deltaY
	return driveResult{location: &Coordinate{Lat: lat, Lon: lon}, lastWp: wp, distanceBetween: distanceFromLast + currentDistance}
}
