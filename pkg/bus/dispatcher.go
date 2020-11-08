package bus

import (
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"time"
)

type Dispatcher struct {
	busModel model.BusModel
	gps      model.RouteService
	publish  model.Publisher
}

func NewDispatcher(mdl model.BusModel, publisher model.Publisher, routeService model.RouteService) *Dispatcher {
	return &Dispatcher{busModel: mdl, publish: publisher, gps: routeService}
}

func (d *Dispatcher) Start(start model.Time) {
	for _, modelBus := range d.busModel.Buses() {
		bus := bus{id: modelBus.Id, assignments: modelBus.Assignments, gps: d.gps, dispatcher: d, position: modelBus.Assignments[0].WayPoints[0]}
		timer := model.NewTicker(start, 500*time.Millisecond, 500*time.Millisecond)
		go bus.start(timer)
	}
}

func (d *Dispatcher) positionStatement(bus *bus, current model.Coordinate) {
	d.publish(model.BusPosition{BusId: bus.id, Location: [2]float64{current.Lat(), current.Lon()}})
}
