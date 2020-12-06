package bus

import (
	"fmt"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"sync"
)

// Dispatcher orchestrates all bus movements in the system. A Dispatcher should always
// be created with NewDispatcher.
type Dispatcher struct {
	busModel    model.BusModel
	buses       map[model.BusId]*bus
	gps         model.RouteService
	publish     model.Publisher
	Frequency   float64
	Warp        float64
	BusSpeedKmh int
}

// NewDispatcher creates a dispatcher with the given parameters.
func NewDispatcher(mdl model.BusModel, publisher model.Publisher, routeService model.RouteService) *Dispatcher {
	return &Dispatcher{busModel: mdl, publish: publisher, gps: routeService, Frequency: 2, Warp: 1, BusSpeedKmh: 40, buses: make(map[model.BusId]*bus)}
}

// Run starts all busses the dispatcher is aware of. This method blocks until all buses have finished all their assignments.
func (d *Dispatcher) Run(start model.Time) {
	var wg sync.WaitGroup
	for _, modelBus := range d.busModel.Buses() {
		bus := bus{id: modelBus.Id, assignments: modelBus.Assignments, gps: d.gps, dispatcher: d, position: modelBus.Assignments[0].WayPoints[0], speed: d.BusSpeedKmh}
		timer := model.NewTicker(start, d.Frequency, d.Warp)
		wg.Add(1)
		d.buses[bus.id] = &bus
		go func() {
			defer wg.Done()
			defer timer.Stop()
			bus.start(timer)
		}()
	}
	wg.Wait()
}

func (d *Dispatcher) positionStatement(bus *bus, current model.Coordinate) {
	d.publish(model.BusPosition{BusId: bus.id, Location: [2]float64{current.Lat(), current.Lon()}})
}

// QueryCurrentAssignment gets the current assignment with the bus with the given id. If the bus with the
// id does not exist, this method will panic. Callers of this method should know which buses the dispatcher contains.
func (d *Dispatcher) QueryCurrentAssignment(id model.BusId) *model.Assignment {
	bus, ok := d.buses[id]
	if !ok {
		panic(fmt.Sprintf("bus with busId \"%s\"not found", id))
	}
	return bus.getCurrentAssignment()
}
