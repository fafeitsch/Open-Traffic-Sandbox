package route

import (
	"fmt"
	"testing"
)

func TestRoutedVehicle_drive(t *testing.T) {
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

	vehicle := RoutedVehicle{Waypoints: &cc1, SpeedKmh: 50, Id: 4242}

	actual := vehicle.drive(&cc1, 175, 550)
	expected := driveResult{location: &Coordinate{Lat: 6.875, Lon: 2.375}, lastWp: &cc2, distanceBetween: 225}
	compareDriveResult(expected, actual, t)

	actual = vehicle.drive(&cc2, 225, 800)
	expected = driveResult{location: &Coordinate{Lat: 11.25, Lon: 4.625}, lastWp: &cc4, distanceBetween: 25}
	compareDriveResult(expected, actual, t)

	actual = vehicle.drive(&cc2, 225, 875)
	expected = driveResult{location: &Coordinate{Lat: 13.5, Lon: 3.5}, lastWp: &cc5, distanceBetween: 0}
	compareDriveResult(expected, actual, t)

	actual = vehicle.drive(&cc2, 225, 1000)
	expected = driveResult{location: &Coordinate{Lat: 13.5, Lon: 3.5}, lastWp: &cc5, distanceBetween: 0}
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
