package bus

import (
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"log"
	"math"
)

type bus struct {
	id             model.BusId
	dispatcher     *Dispatcher
	assignments    []model.Assignment
	gps            model.RouteService
	heartBeatTimer model.Ticker
	position       model.Coordinate
	speed          int
}

func (b *bus) start(timer model.Ticker) {
	b.heartBeatTimer = timer
	for _, assignment := range b.assignments {
		b.handleAssignment(assignment)
	}
}

func (b *bus) handleAssignment(a model.Assignment) {
	last := <-b.heartBeatTimer.HeartBeat
	for last.Before(a.Departure) {
		current, ok := <-b.heartBeatTimer.HeartBeat
		if !ok {
			return
		}
		last = current
	}

	for _, wayPoint := range a.WayPoints {
		route, _, err := b.gps(b.position, &wayPoint)
		if err != nil {
			log.Printf("bus %s: could not find route, skipping route: %v", b.id, err)
		}
		for len(route) > 0 {
			current, ok := <-b.heartBeatTimer.HeartBeat
			if !ok {
				return
			}
			deltaTime := current.Sub(last).Seconds()
			driven := (float64(b.speed) / 3.6) * deltaTime
			route = b.drive(route, driven)
			last = current
		}
	}
}

func (b *bus) drive(route []model.Coordinate, distanceToDrive float64) []model.Coordinate {
	defer b.dispatcher.positionStatement(b, b.position)
	if b.position == route[0] {
		route = route[1:]
	}
	distanceToNext := distanceTo(b.position, route[0])
	newRoute := route
	for distanceToDrive >= distanceToNext && len(newRoute) > 1 {
		distanceToDrive = distanceToDrive - distanceToNext
		distanceToNext = distanceTo(newRoute[0], newRoute[1])
		b.position = &coordinate{newRoute[0].Lat(), newRoute[0].Lon()}
		newRoute = newRoute[1:]
	}
	if distanceToDrive >= distanceToNext {
		b.position = &coordinate{newRoute[0].Lat(), newRoute[0].Lon()}
		return []model.Coordinate{}
	}
	lambda := distanceToDrive / distanceToNext
	deltaX := newRoute[0].Lat() - b.position.Lat()
	deltaY := newRoute[0].Lon() - b.position.Lon()
	lat := b.position.Lat() + lambda*deltaX
	lon := b.position.Lon() + lambda*deltaY
	b.position = &coordinate{lat: lat, lon: lon}
	return newRoute
}

type coordinate struct {
	lat float64
	lon float64
}

func (c *coordinate) Lat() float64 {
	return c.lat
}

func (c *coordinate) Lon() float64 {
	return c.lon
}

func distanceTo(c model.Coordinate, other model.Coordinate) float64 {
	earthRadius := 6371000.0 // meters
	delta1 := toRadians(c.Lat())
	delta2 := toRadians(other.Lat())
	deltaPhi := toRadians(other.Lat() - c.Lat())
	deltaLambda := toRadians(other.Lon() - c.Lon())
	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(delta1)*math.Cos(delta2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	atan := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * atan
}

func toRadians(degree float64) float64 {
	return degree * (math.Pi / 180)
}
