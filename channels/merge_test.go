package channels

import (
	"github.com/fafeitsch/Open-Traffic-Sandbox/routing"
	"strconv"
	"sync"
	"testing"
)

func TestMerge(t *testing.T) {
	locations := generateVehicleLocations()
	expectedLocations := make(map[string]bool)
	for _, location := range locations {
		if _, ok := expectedLocations[location.VehicleId]; ok {
			t.Errorf("The sample locations contains at least one duplicate vehicle id: %s", location.VehicleId)
		}
		expectedLocations[location.VehicleId] = true
	}
	channels := make([]<-chan routing.VehicleLocation, 0)
	for i := 0; i < len(locations)/10; i++ {
		channel := make(chan routing.VehicleLocation)
		go func(index int) {
			for counter := 0; counter < 10; counter++ {
				channel <- locations[index*10+counter]
			}
			close(channel)
		}(i)
		channels = append(channels, channel)
	}
	mainChannel := Merge(channels)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		loc, ok := <-mainChannel
		for ok {
			expectedLocations[loc.VehicleId] = false
			loc, ok = <-mainChannel
		}
		wg.Done()
	}()
	wg.Wait()
	for index, location := range locations {
		if expectedLocations[location.VehicleId] {
			t.Errorf("The vehicle with ID %s at index %d was never reveived!", location.VehicleId, index)
		}
	}
}

func generateVehicleLocations() []routing.VehicleLocation {
	result := make([]routing.VehicleLocation, 0)
	for i := 0; i < 60; i++ {
		vehicleLocation := routing.VehicleLocation{VehicleId: strconv.Itoa(i), Location: [2]float64{0, 0}}
		result = append(result, vehicleLocation)
	}
	return result
}
