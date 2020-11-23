package bus

import (
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

type mockModel struct {
	buses []model.Bus
}

func (m mockModel) Buses() []model.Bus {
	return m.buses
}

func TestDispatcher_Start(t *testing.T) {
	// This test is somewhat artificial. We are modelling one bus who has a slice of way points.
	// As route function, we simply return the start and the end coordinate, meaning that the route between to coordinates
	// is a simple line. In reality, OSRM returns are really detailed path instead of a simple line.
	bus1 := model.Bus{
		Id: "Bus1",
		Assignments: []model.Assignment{
			{
				Name:      "test assignment1",
				Departure: model.MustParseTime("15:00"),
				WayPoints: []model.WayPoint{
					{Longitude: 9.95075, Latitude: 49.79993}, // https://www.openstreetmap.org/?mlon=9.95075&mlat=49.79993#map=16/49.8020/9.9519
					{Longitude: 9.94932, Latitude: 49.79900}, // https://www.openstreetmap.org/?mlon=9.94932&mlat=49.79900#map=15/49.7980/9.9446
					{Longitude: 9.94550, Latitude: 49.79886}, // https://www.openstreetmap.org/?mlon=9.94550&mlat=49.79886#map=17/49.79848/9.94700
					{Longitude: 9.94449, Latitude: 49.79871}, // https://www.openstreetmap.org/?mlon=9.94449&mlat=49.79871#map=16/49.7963/9.9464
					{Longitude: 9.94316, Latitude: 49.79919}, // https://www.openstreetmap.org/?mlon=9.94316&mlat=49.79919#map=18/49.79858/9.94369
				},
			},
		},
	}
	positions := make([]model.BusPosition, 0, 0)
	positionReceiver := make(chan model.BusPosition)
	publisher := func(position model.BusPosition) {
		positionReceiver <- position
	}
	routeService := func(coordinates ...model.Coordinate) ([]model.Coordinate, float64, error) {
		return coordinates, 0, nil
	}
	dispatcher := NewDispatcher(&mockModel{buses: []model.Bus{bus1}}, publisher, routeService)
	dispatcher.Frequency = 1000
	dispatcher.Warp = 10000
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for position := range positionReceiver {
			positions = append(positions, position)
		}
	}()
	dispatcher.Run(model.MustParseTime("16:49"))
	close(positionReceiver)
	wg.Wait()
	// With the current speed, we have 111.111 meters per 10 seconds. However, a bus
	// cannot finish one route and start the next in one tick. Thus, if there current route still has
	// 80 meters, but the tick requires 111 meters to drive, then 31 meters are lost in that tick (the bus
	// essentially slows down). Therefore, we need 8 ticks to cover the 603 meters. (in reality 650 meters,
	// but we are working with the assumption that the earth is a perfect sphere.
	assert.Equal(t, 8, len(positions))
}
