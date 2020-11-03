package bus

import (
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"time"
)

type Dispatcher struct {
	busModel  model.BusModel
	gps       model.RouteService
	publisher model.Publisher
}

func NewDispatcher(mdl model.BusModel) *Dispatcher {
	return &Dispatcher{busModel: mdl}
}

func (d *Dispatcher) Start(start time.Time) {
	for _, modelBus := range d.busModel.Buses() {
		bus := bus{id: modelBus.Id, assignments: modelBus.Assignments, gps: d.gps}
		bus.ready(start)
	}
}

func (d *Dispatcher) positionStatement(bus *bus, current model.Coordinate) {
	d.publisher(model.BusPosition{BusId: bus.id, Location: [2]float64{current.Lat(), current.Lon()}})
}
