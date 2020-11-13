package bus

import (
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"sync"
	"time"
)

// Dispatcher orchestrates all bus movements in the system. A Dispatcher should always
// be created with NewDispatcher.
type Dispatcher struct {
	busModel       model.BusModel
	gps            model.RouteService
	publish        model.Publisher
	RealtimeTick   time.Duration
	SimulationTick time.Duration
}

// NewDispatcher creates a dispatcher with the given parameters.
func NewDispatcher(mdl model.BusModel, publisher model.Publisher, routeService model.RouteService) *Dispatcher {
	return &Dispatcher{busModel: mdl, publish: publisher, gps: routeService, RealtimeTick: 500 * time.Millisecond, SimulationTick: 500 * time.Millisecond}
}

// Run starts all busses the dispatcher is aware of. This method blocks until all buses have finished all their assignments.
func (d *Dispatcher) Run(start model.Time) {
	var wg sync.WaitGroup
	for _, modelBus := range d.busModel.Buses() {
		bus := bus{id: modelBus.Id, assignments: modelBus.Assignments, gps: d.gps, dispatcher: d, position: modelBus.Assignments[0].WayPoints[0]}
		timer := model.NewTicker(start, d.SimulationTick, d.RealtimeTick)
		wg.Add(1)
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
