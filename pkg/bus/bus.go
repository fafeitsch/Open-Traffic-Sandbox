package bus

import (
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"log"
	"math"
	"sync"
)

type bus struct {
	mutex             sync.Mutex
	id                model.BusId
	dispatcher        *Dispatcher
	assignments       []model.Assignment
	currentAssignment int
	gps               model.RouteService
	heartBeatTimer    model.Ticker
	position          model.Coordinate
	speed             int
	currentStop       *model.WayPoint
}

func (b *bus) start(timer model.Ticker) {
	b.heartBeatTimer = timer
	for index, assignment := range b.assignments {
		b.setAssignmentIndex(index)
		b.handleAssignment(assignment)
	}
}

func (b *bus) setAssignmentIndex(index int) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.currentAssignment = index
}

func (b *bus) getCurrentAssignment() *model.Assignment {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return &b.assignments[b.currentAssignment]
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
		if wayPoint.Id != nil {
			b.currentStop = &wayPoint
		}
		for last.Before(wayPoint.Departure) {
			b.dispatcher.publish(model.BusPosition{Stop: b.currentStop, BusId: b.id, Location: [2]float64{b.position.Lat(), b.position.Lon()}})
			current, ok := <-b.heartBeatTimer.HeartBeat
			if !ok {
				return
			}
			last = current
		}
		b.currentStop = nil
	}
}

func (b *bus) drive(route []model.Coordinate, distanceToDrive float64) []model.Coordinate {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	defer b.dispatcher.publish(model.BusPosition{BusId: b.id, Location: [2]float64{b.position.Lat(), b.position.Lon()}})
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
