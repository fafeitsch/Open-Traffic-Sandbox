package domain

import (
	"fmt"
	"math"
	"time"
)

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// PolylineEqual returns true if both coordinates are equally encoded using
// the polyline format. With other words, both coordinates must be identical
// in their first five positions after the decimal dot.
func (c *Coordinate) PolylineEqual(other Coordinate) bool {
	latDiff := math.Abs(c.Lat - other.Lat)
	lonDiff := math.Abs(c.Lon - other.Lon)
	return latDiff < 0.00001 && lonDiff < 0.00001
}

type Coordinates []Coordinate

func (c Coordinates) chain() *ChainedCoordinate {
	geometry := c
	cumulatedDistance := 0.0
	max := 0.0
	firstChainedCoordinate := ChainedCoordinate{Coordinate: geometry[0]}
	coordinate := &firstChainedCoordinate
	for _, c := range geometry[1:] {
		nextChainedCoordinate := ChainedCoordinate{Coordinate: c}
		coordinate.DistanceToNext = coordinate.distanceTo(&c)
		cumulatedDistance = cumulatedDistance + coordinate.DistanceToNext
		coordinate.Next = &nextChainedCoordinate
		if coordinate.DistanceToNext > max {
			max = coordinate.DistanceToNext
		}
		coordinate = coordinate.Next
	}
	return &firstChainedCoordinate
}

func (c *Coordinate) distanceTo(other *Coordinate) float64 {
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

type VehicleLocation struct {
	Location  [2]float64 `json:"loc"`
	VehicleId string     `json:"id"`
}

type Assignment struct {
	Start              time.Time
	Waypoints          Coordinates
	precomputed        *ChainedCoordinate
	destinationHandler func(time.Time, *Vehicle, *Assignment)
}

func (a *Assignment) activate() driveResult {
	if a.precomputed == nil {
		a.precomputed = a.Waypoints.chain()
	}
	first := a.precomputed
	return driveResult{location: &first.Coordinate, lastWp: first, distanceBetween: 0, destinationReached: false}
}

func (a *Assignment) destinationReached(now time.Time, v *Vehicle, next *Assignment) {
	if a.destinationHandler == nil {
		return
	}
	a.destinationHandler(now, v, next)
}

type Vehicle struct {
	SpeedKmh         float64
	Id               string
	HeartBeat        <-chan time.Time
	Assignments      []Assignment
	activeAssignment Assignment
}

func (v *Vehicle) StartJourney(consumer chan<- VehicleLocation) {
	if len(v.Assignments) == 0 {
		return
	}
	speedMS := v.SpeedKmh / 3.6
	v.activeAssignment = v.Assignments[0]
	driveResult := v.activeAssignment.activate()
	v.Assignments = v.Assignments[1:]
	last := <-v.HeartBeat
	consumer <- VehicleLocation{Location: [2]float64{v.activeAssignment.precomputed.Lat, v.activeAssignment.precomputed.Lon}, VehicleId: v.Id}
	for {
		now, ok := <-v.HeartBeat
		if !ok {
			close(consumer)
			return
		}
		if v.activeAssignment.Start.After(now) {
			last = now
			continue
		}
		deltaTime := now.Sub(last).Seconds()
		driven := speedMS * deltaTime
		driveResult = v.drive(driveResult.lastWp, driveResult.distanceBetween, driven)
		last = now
		location := VehicleLocation{Location: [2]float64{driveResult.location.Lat, driveResult.location.Lon}, VehicleId: v.Id}
		consumer <- location
		if driveResult.destinationReached {
			if len(v.Assignments) == 0 {
				v.activeAssignment.destinationReached(now, v, nil)
				close(consumer)
				return
			}
			v.activeAssignment.destinationReached(now, v, &v.Assignments[0])
			v.activeAssignment = v.Assignments[0]
			driveResult = v.activeAssignment.activate()
			v.Assignments = v.Assignments[1:]
		}
	}
}

type driveResult struct {
	location           *Coordinate
	lastWp             *ChainedCoordinate
	distanceBetween    float64
	destinationReached bool
}

func (v *Vehicle) drive(lastWp *ChainedCoordinate, distanceBetween float64, distanceToDrive float64) driveResult {
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
