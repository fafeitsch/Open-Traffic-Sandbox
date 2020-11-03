package bus

import (
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"time"
)

const speedMs = 40 / 3.6

type bus struct {
	id             model.BusId
	dispatcher     Dispatcher
	assignments    []model.Assignment
	gps            model.RouteService
	heartBeatTimer *time.Timer
	position	   model.Coordinate
}

func (b *bus) ready(start time.Time) {
	b.heartBeatTimer = time.NewTimer(1 * time.Second)
	for _, assignment := range b.assignments {

	}
}

func (b *bus) handleAssignment(a model.Assignment) {
	destinationReached := false
	last := <-b.heartBeatTimer.C
	for !destinationReached {
		current, ok := <- b.heartBeatTimer.C
		if !ok {
			return
		}
		if current.Before(a.Departure.inter)
	}
}
