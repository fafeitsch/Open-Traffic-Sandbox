package routing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"testing"
	"time"
)

func TestPointsToCoordinates(t *testing.T) {
	coords := [][]float64{{1, 2}, {3, 4}, {5, 6}}
	actual := PointsToCoordinates(coords)
	if len(actual) != len(coords) {
		t.Errorf("Expectd length %d, but was %d", len(coords), len(actual))
	}
	for index, coord := range coords {
		c := actual[index]
		if c.Lat != coord[0] {
			t.Errorf("At index %d, expected latitude is %f, but was %f", index, coord[0], c.Lat)
		}
		if c.Lon != coord[1] {
			t.Errorf("At index %d, expected longitude is %f, but was %f", index, coord[0], c.Lat)
		}
	}
}

func TestCoordinate_DistanceTo(t *testing.T) {
	c1 := Coordinate{Lat: 49.792778, Lon: 9.938611}
	c2 := Coordinate{Lat: 49.801389, Lon: 9.935556}
	expected := 982.2866513649033
	actual := c1.DistanceTo(&c2)
	if expected != actual {
		t.Errorf("Expected %f, actual %f", expected, actual)
	}
	if actual != c2.DistanceTo(&c1) {
		t.Errorf("Distance is not symmetric")
	}
	if c1.DistanceTo(&c1) != 0 {
		t.Errorf("Distance to the same coordinate should be zero")
	}
}

func TestRoutedVehicle_StartJourney(t *testing.T) {
	receiver := make(chan VehicleLocation, 1)
	_, _, _, _, vehicle := createSampleVehicle()
	ticker := make(chan time.Time)
	go func() {
		now := time.Now()
		for i := 0; i < 3000; i++ {
			ticker <- now
			now = now.Add(40 * time.Millisecond)
		}
		close(ticker)
	}()
	vehicle.HeartBeat = ticker
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, open := <-receiver
		events := 1
		for open {
			_, open = <-receiver
			if open {
				events = events + 1
			}
		}
		if events != 2882 {
			t.Errorf("2882 events should be generated but there were %d events.", events)
		}
	}()
	vehicle.StartJourney(receiver)
	wg.Wait()
}

func readWaypoints(filename string) Coordinates {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("could not read test resource: %v", err)
	}
	var array [][]float64
	err = json.Unmarshal(data, &array)
	if err != nil {
		log.Fatalf("could not parse test resource: %v", err)
	}
	result := make([]Coordinate, 0)
	for _, coord := range array {
		result = append(result, Coordinate{Lat: coord[0], Lon: coord[1]})
	}
	return result

}

var expectedLocations = []Coordinate{
	{Lat: 49.800090, Lon: 9.984000},
}

func TestRoutedVehicle_StartJourney2(t *testing.T) {
	receiver := make(chan VehicleLocation, 1)
	car := RoutedVehicle{
		Assignments: []Assignment{
			{Line: &Line{Waypoints: readWaypoints("testdata/sampleRoute.json")}},
		},
		SpeedKmh: 50,
	}
	ticker := make(chan time.Time)
	go func() {
		now := time.Now()
		for i := 0; i < 3000; i++ {
			ticker <- now
			now = now.Add(40 * time.Millisecond)
		}
		close(ticker)
	}()
	car.HeartBeat = ticker
	var wg sync.WaitGroup
	counter := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		location, open := <-receiver
		if counter < len(expectedLocations) {
			err := assertSameCoordinate(&expectedLocations[counter], newCoordinate(location.Location))
			if err != nil {
				t.Errorf("Error in step %d: %v", counter, err)
			}
		}
		for open {
			counter = counter + 1
			location, o := <-receiver
			if o {
				if counter < len(expectedLocations) {
					err := assertSameCoordinate(&expectedLocations[counter], newCoordinate(location.Location))
					if err != nil {
						t.Errorf("Error in step %d: %v", counter, err)
					}
				}
				counter = counter + 1
			}
			open = o
		}
	}()
	car.StartJourney(receiver)
	wg.Wait()
}

func createSampleVehicle() (ChainedCoordinate, ChainedCoordinate, ChainedCoordinate, ChainedCoordinate, RoutedVehicle) {
	c1_1 := Coordinate{Lat: 1, Lon: 1}
	c35_35 := Coordinate{Lat: 3.5, Lon: 3.5}
	c8_2 := Coordinate{Lat: 8, Lon: 2}
	c105_5 := Coordinate{Lat: 10.5, Lon: 5}
	c135_35 := Coordinate{Lat: 13.5, Lon: 3.5}
	cc5 := ChainedCoordinate{Coordinate: c135_35, Next: nil, DistanceToNext: 0}
	cc4 := ChainedCoordinate{Coordinate: c105_5, Next: &cc5, DistanceToNext: 100}
	cc3 := ChainedCoordinate{Coordinate: c8_2, Next: &cc4, DistanceToNext: 700}
	cc2 := ChainedCoordinate{Coordinate: c35_35, Next: &cc3, DistanceToNext: 300}
	cc1 := ChainedCoordinate{Coordinate: c1_1, Next: &cc2, DistanceToNext: 500}
	vehicle := RoutedVehicle{
		Assignments: []Assignment{
			{precomputed: &cc1},
		},
		SpeedKmh: 50,
		Id:       "4242"}
	return cc5, cc4, cc2, cc1, vehicle
}

func TestRoutedVehicle_drive(t *testing.T) {
	cc5, cc4, cc2, cc1, vehicle := createSampleVehicle()

	actual := vehicle.drive(&cc1, 175, 550)
	expected := driveResult{location: &Coordinate{Lat: 6.875, Lon: 2.375}, lastWp: &cc2, distanceBetween: 225}
	compareDriveResult(expected, actual, t)

	actual = vehicle.drive(&cc2, 225, 800)
	expected = driveResult{location: &Coordinate{Lat: 11.25, Lon: 4.625}, lastWp: &cc4, distanceBetween: 25}
	compareDriveResult(expected, actual, t)

	actual = vehicle.drive(&cc2, 225, 875)
	expected = driveResult{location: &Coordinate{Lat: 13.5, Lon: 3.5}, lastWp: &cc5, distanceBetween: 0, destinationReached: true}
	compareDriveResult(expected, actual, t)

	actual = vehicle.drive(&cc2, 225, 1000)
	expected = driveResult{location: &Coordinate{Lat: 13.5, Lon: 3.5}, lastWp: &cc5, distanceBetween: 0, destinationReached: true}
	compareDriveResult(expected, actual, t)
}

func compareDriveResult(expected driveResult, actual driveResult, t *testing.T) {
	if err := assertSameChainedCoordinate(expected.lastWp, actual.lastWp); err != nil {
		t.Errorf("%v", err)
	}
	if err := assertSameCoordinate(expected.location, actual.location); err != nil {
		t.Errorf("%v", err)
	}
	if expected.distanceBetween != actual.distanceBetween {
		t.Errorf("Expected distanceBetween %f; Actual distanceBetween %f", expected.distanceBetween, actual.distanceBetween)
	}
	if expected.destinationReached != actual.destinationReached {
		t.Errorf("Expected destinationReached %t; Actual destinationReached %t", expected.destinationReached, actual.destinationReached)
	}
}

func assertSameCoordinate(expected *Coordinate, actual *Coordinate) error {
	if expected.Lat != actual.Lat {
		return fmt.Errorf("expected latitude %f; actual latitude %f", expected.Lat, actual.Lat)
	}
	if expected.Lon != actual.Lon {
		return fmt.Errorf("expected longitude %f; actual longitude %f", expected.Lon, actual.Lon)
	}
	return nil
}

func assertSameChainedCoordinate(expected *ChainedCoordinate, actual *ChainedCoordinate) error {
	if expected.Lat != actual.Lat {
		return fmt.Errorf("expected latitude %f; actual latitude %f", expected.Lat, actual.Lat)
	}
	if expected.Lon != actual.Lon {
		return fmt.Errorf("expected longitude %f; actual longitude %f", expected.Lon, actual.Lon)
	}
	if expected.DistanceToNext != actual.DistanceToNext {
		return fmt.Errorf("expected distanceToNext %f; actual distanceToNext %f", expected.DistanceToNext, actual.DistanceToNext)
	}
	if expected.Next != nil && actual.Next == nil {
		return fmt.Errorf("The next waypoint should be nil, but has a value")
	}
	if expected.Next == nil && actual.Next != nil {
		return fmt.Errorf("The next waypoint should have a predecessor, but has not")
	}
	if expected.Next != nil && actual.Next != nil {
		return assertSameChainedCoordinate(expected.Next, actual.Next)
	}
	return nil
}
