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

type Coordinates []Coordinate

func (c Coordinates) Chain() *ChainedCoordinate {
	geometry := c
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
	return &firstChainedCoordinate
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

type Line struct {
	Id        string
	Name      string
	Waypoints Coordinates
}

type Assignment struct {
	Start       time.Time
	Line        *Line
	StartFrom   *Coordinate
	GoTo        *Coordinate
	precomputed *ChainedCoordinate
}

func (a *Assignment) activate() activeAssignment {
	if a.Line != nil {
		return activeAssignment{start: a.Start, waypoints: a.Line.Waypoints.Chain()}
	} else if a.StartFrom != nil {
		return activeAssignment{start: a.Start, waypoints: Coordinates([]Coordinate{*a.GoTo}).Chain()}
	} else if a.precomputed != nil {
		return activeAssignment{start: a.Start, waypoints: a.precomputed}
	}
	panic("assignment is invalid, either line, startFrom, or precomputed must be != nil")
}

type activeAssignment struct {
	start     time.Time
	waypoints *ChainedCoordinate
}

type RoutedVehicle struct {
	SpeedKmh         float64
	Id               string
	HeartBeat        <-chan time.Time
	Assignments      []Assignment
	activeAssignment activeAssignment
}

func (v *RoutedVehicle) StartJourney(consumer chan<- VehicleLocation) {
	if len(v.Assignments) == 0 {
		return
	}
	speedMS := v.SpeedKmh / 3.6
	v.activeAssignment = v.Assignments[0].activate()
	v.Assignments = v.Assignments[1:]
	driveResult := createEmptyResult(v.activeAssignment.waypoints)
	last := <-v.HeartBeat
	consumer <- VehicleLocation{Location: [2]float64{v.activeAssignment.waypoints.Lat, v.activeAssignment.waypoints.Lon}, VehicleId: v.Id}
	for {
		now, ok := <-v.HeartBeat
		if !ok {
			close(consumer)
			return
		}
		if v.activeAssignment.start.After(now) {
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
				close(consumer)
				return
			}
			v.activeAssignment = v.Assignments[0].activate()
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
