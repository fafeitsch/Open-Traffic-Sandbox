package routing

import (
	"encoding/json"
	"fmt"
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
			t.Errorf("2216 events should be generated but there were %d events.", events)
		}
	}()
	vehicle.StartJourney(receiver)
	wg.Wait()
}

func staticRouteService([]Coordinate) ([]Coordinate, float64, error) {
	s := `[[49.800090, 9.984000],[49.800090, 9.984180],[49.800350, 9.985120],[49.800410, 9.985500],[49.800610, 9.986560],[49.800980, 9.985750],[49.801160, 9.985390],[49.801260, 9.985410],[49.801470, 9.985580],[49.801730, 9.985890],[49.801950, 9.985420],[49.802140, 9.984930],[49.802190, 9.984690],[49.802210, 9.984570],[49.802210, 9.984450],[49.802180, 9.984150],[49.802110, 9.983720],[49.802010, 9.983090],[49.801980, 9.982920],[49.801880, 9.982260],[49.801790, 9.981760],[49.801680, 9.981350],[49.801530, 9.980880],[49.801490, 9.980710],[49.801470, 9.980540],[49.801470, 9.980370],[49.801480, 9.980240],[49.801510, 9.980120],[49.801580, 9.979940],[49.801610, 9.979860],[49.801730, 9.979590],[49.801750, 9.979520],[49.801770, 9.979400],[49.801800, 9.979400],[49.801820, 9.979380],[49.801840, 9.979360],[49.801850, 9.979330],[49.801860, 9.979300],[49.801860, 9.979260],[49.801850, 9.979230],[49.801840, 9.979200],[49.801820, 9.979170],[49.801800, 9.979160],[49.801770, 9.979150],[49.801750, 9.979160],[49.801730, 9.979180],[49.801710, 9.979210],[49.801700, 9.979240],[49.801600, 9.979200],[49.801520, 9.979130],[49.801410, 9.978990],[49.800780, 9.978230],[49.800640, 9.978070],[49.800530, 9.978020],[49.800040, 9.977790],[49.800090, 9.977500],[49.800220, 9.976890],[49.800260, 9.976230],[49.800260, 9.975880],[49.800250, 9.975420],[49.800110, 9.973640],[49.800070, 9.973230],[49.799910, 9.971220],[49.799870, 9.971070],[49.799810, 9.970970],[49.799400, 9.970930],[49.799220, 9.970930],[49.798790, 9.970930],[49.798690, 9.970940],[49.797950, 9.971030],[49.797870, 9.971050],[49.797840, 9.971050],[49.797790, 9.971060],[49.797550, 9.971100],[49.797150, 9.971170],[49.796610, 9.971270],[49.796550, 9.971280],[49.796520, 9.971280],[49.796470, 9.971290],[49.796050, 9.971390],[49.795980, 9.971410],[49.795890, 9.971410],[49.795880, 9.970970],[49.795880, 9.970360],[49.795950, 9.969160],[49.796010, 9.967780],[49.796080, 9.966820],[49.796110, 9.966480],[49.796170, 9.965880],[49.796180, 9.965760],[49.796230, 9.965280],[49.796270, 9.964850],[49.796030, 9.964760],[49.795770, 9.964640],[49.795740, 9.964600],[49.795720, 9.964500],[49.795750, 9.964020],[49.795790, 9.963570],[49.795840, 9.963180],[49.796230, 9.959400],[49.796270, 9.959030],[49.796280, 9.958930],[49.796310, 9.958690],[49.796310, 9.958660],[49.796400, 9.957810],[49.796430, 9.957520],[49.796470, 9.957120],[49.796490, 9.956940],[49.796530, 9.956580],[49.796580, 9.956190],[49.796620, 9.955790],[49.796660, 9.955430],[49.796670, 9.955290],[49.796680, 9.954960],[49.796660, 9.954600],[49.796660, 9.954330],[49.796690, 9.954020],[49.796780, 9.953430],[49.796810, 9.953220],[49.796830, 9.953140],[49.796860, 9.952930],[49.796880, 9.952790],[49.796900, 9.952590],[49.796970, 9.952280],[49.797080, 9.952000],[49.797170, 9.951890],[49.797240, 9.951820],[49.797330, 9.951770],[49.797440, 9.951760],[49.797550, 9.951780],[49.797630, 9.951830],[49.797730, 9.951910],[49.797790, 9.952000],[49.797840, 9.952110],[49.797890, 9.952230],[49.797900, 9.952330],[49.797920, 9.952440],[49.797920, 9.952580],[49.797890, 9.952690],[49.797850, 9.952780],[49.797780, 9.952890],[49.797710, 9.952940],[49.797620, 9.952950],[49.797550, 9.952910],[49.797340, 9.952740],[49.797120, 9.952480],[49.796910, 9.952280],[49.796800, 9.952180],[49.796700, 9.952110],[49.796480, 9.951980],[49.796290, 9.951900],[49.796110, 9.951830],[49.795520, 9.951630],[49.795370, 9.951570],[49.795220, 9.951510],[49.795120, 9.951470],[49.794940, 9.951390],[49.794750, 9.951280],[49.794190, 9.950900],[49.794040, 9.950790],[49.793870, 9.950660],[49.793570, 9.950460],[49.793300, 9.950250],[49.792410, 9.949290],[49.792280, 9.949150],[49.791810, 9.948690],[49.790610, 9.947520],[49.789630, 9.946230],[49.789520, 9.946080],[49.789460, 9.946000],[49.788550, 9.944480],[49.788170, 9.943900],[49.787990, 9.943630],[49.787830, 9.943360],[49.787680, 9.943130],[49.787470, 9.942820],[49.787270, 9.942530],[49.787040, 9.942240],[49.786850, 9.942040],[49.786690, 9.941890],[49.786490, 9.941710],[49.786270, 9.941520],[49.786070, 9.941370],[49.785880, 9.941240],[49.785690, 9.941110],[49.785520, 9.940980],[49.785360, 9.940890],[49.785010, 9.940740],[49.784600, 9.940630],[49.784190, 9.940610],[49.783640, 9.940630],[49.783400, 9.940680],[49.783240, 9.940730],[49.782420, 9.940980],[49.781990, 9.941110],[49.780330, 9.941620],[49.779460, 9.941910],[49.778710, 9.942160],[49.778350, 9.942310],[49.777920, 9.942540],[49.777530, 9.942760],[49.777460, 9.942800],[49.777370, 9.942830],[49.777250, 9.942840],[49.777150, 9.942830],[49.777040, 9.942810],[49.776940, 9.942760],[49.776850, 9.942710],[49.776680, 9.942600],[49.776560, 9.942520],[49.776450, 9.942460],[49.776320, 9.942420],[49.776150, 9.942420],[49.775990, 9.942450],[49.775630, 9.942560],[49.775430, 9.942640],[49.775160, 9.942730],[49.775030, 9.942770],[49.774790, 9.942810],[49.774570, 9.942830],[49.774280, 9.942830],[49.774130, 9.942780],[49.774010, 9.942720],[49.773900, 9.942630],[49.773780, 9.942540],[49.773650, 9.942430],[49.773300, 9.942060],[49.772350, 9.941020],[49.772330, 9.941000],[49.772070, 9.940730],[49.771910, 9.940560],[49.771380, 9.940010],[49.770490, 9.939120],[49.770370, 9.939010],[49.769490, 9.938180],[49.769100, 9.937830],[49.768850, 9.937610],[49.768550, 9.937310],[49.768070, 9.936890],[49.767910, 9.936570],[49.767860, 9.936420],[49.767850, 9.936320],[49.767850, 9.936200],[49.767870, 9.936070],[49.767910, 9.935960],[49.767970, 9.935880],[49.768060, 9.935790],[49.768120, 9.935730],[49.768220, 9.935740],[49.768330, 9.935770],[49.768430, 9.935840],[49.768490, 9.935900],[49.768570, 9.936000],[49.768710, 9.936190],[49.768820, 9.936310],[49.769070, 9.936590],[49.769170, 9.936710],[49.769250, 9.936810],[49.769310, 9.936940],[49.769620, 9.936510],[49.770150, 9.935700],[49.770790, 9.934800],[49.771250, 9.934130],[49.771990, 9.933070],[49.772280, 9.932650],[49.772870, 9.931820],[49.774180, 9.929300],[49.774450, 9.928800],[49.774720, 9.928470],[49.774810, 9.928390],[49.774900, 9.928310],[49.775140, 9.928180],[49.774990, 9.927900],[49.774930, 9.927790],[49.774880, 9.927680],[49.774710, 9.927320],[49.774540, 9.926940],[49.774450, 9.926700],[49.774300, 9.926280],[49.774220, 9.926010],[49.773800, 9.924770],[49.773390, 9.923230],[49.773330, 9.923050],[49.773300, 9.922880],[49.773210, 9.922540],[49.773140, 9.922310],[49.773040, 9.922010],[49.772860, 9.921580],[49.772670, 9.921260],[49.772090, 9.920510],[49.771620, 9.919650],[49.771420, 9.919160],[49.771260, 9.918700],[49.771210, 9.918530],[49.771010, 9.917910],[49.770830, 9.917210],[49.770770, 9.916670],[49.770770, 9.916640],[49.770770, 9.915540],[49.770800, 9.914850],[49.770800, 9.914630],[49.770820, 9.914220],[49.770850, 9.913600],[49.770860, 9.913520],[49.770860, 9.913230],[49.770880, 9.912630],[49.771050, 9.911630],[49.771190, 9.910260],[49.771250, 9.908920],[49.771260, 9.908640],[49.771250, 9.907350],[49.771250, 9.906970],[49.771250, 9.906600],[49.771340, 9.904970],[49.771360, 9.903080],[49.771340, 9.902780],[49.771550, 9.902740],[49.771610, 9.902750],[49.771670, 9.902830],[49.771720, 9.902950],[49.771780, 9.903430],[49.771810, 9.903550],[49.771850, 9.903560],[49.771880, 9.903520],[49.772020, 9.902860],[49.772060, 9.902770],[49.772100, 9.902740],[49.772140, 9.902750],[49.772170, 9.902860],[49.772250, 9.903660]]`
	var array [][]float64
	json.Unmarshal([]byte(s), &array)
	result := make([]Coordinate, 0)
	for _, coord := range array {
		result = append(result, Coordinate{Lat: coord[0], Lon: coord[1]})
	}
	return result, 0, nil
}

var expectedLocations = []Coordinate{
	{Lat: 49.800090, Lon: 9.984000},
}

func TestRoutedVehicle_StartJourney2(t *testing.T) {
	receiver := make(chan VehicleLocation, 1)
	rawVehicle := vehicle{Coordinates: nil, Id: "82"}
	car, _ := rawVehicle.toRoutedVehicle(staticRouteService)
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
	vehicle := RoutedVehicle{Waypoints: &cc1, SpeedKmh: 50, Id: "4242"}
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
