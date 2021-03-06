package model

import "fmt"

func loadBuses(scenario scenario, lines map[LineId]Line) (map[BusId]Bus, error) {
	result := make(map[BusId]Bus)
	for _, scenBus := range scenario.Buses {
		bus := Bus{Id: BusId(scenBus.Id)}
		assignments := make([]Assignment, 0, len(scenBus.Assignments))
		for _, asmgt := range scenBus.Assignments {
			assignment, err := initAssignments(asmgt.Start, asmgt.Line, asmgt.Coordinates, lines)
			if err != nil {
				return nil, fmt.Errorf("could not load bus \"%s\": %v", bus.Id, err)
			}
			assignments = append(assignments, *assignment)
		}
		bus.Assignments = assignments
		result[bus.Id] = bus
	}
	return result, nil
}

func initAssignments(rawStart string, line string, coordinates [][2]float64, lineMap map[LineId]Line) (*Assignment, error) {
	start, err := ParseTime(rawStart)
	if err != nil {
		return nil, fmt.Errorf("could not parse time \"%s\" of bus: %v", rawStart, err)
	}
	if line != "" {
		return createLineAssignment(lineMap, line, start)
	} else {
		return createWaypointAssignment(coordinates, start), nil
	}
}

func createLineAssignment(lineMap map[LineId]Line, rawLine string, start Time) (*Assignment, error) {
	assignment := Assignment{Departure: start}
	line, ok := lineMap[LineId(rawLine)]
	if !ok {
		return nil, fmt.Errorf("line \"%s\" not found", line.Id)
	}
	assignment.Line = &line
	assignment.Name = line.Name
	waypoints := make([]WayPoint, 0, len(line.waypoints))
	departures := line.TourTimes(assignment.Departure)
	if departures == nil {
		return nil, fmt.Errorf("line assignment \"%s\" with start time \"%s\" has no equivalent in time table", line.Id, start)
	}
	index := 0
	for _, wp := range line.waypoints {
		point := WayPoint{Id: wp.Id, Name: wp.Name, Latitude: wp.Latitude, Longitude: wp.Longitude}
		if wp.Id != nil {
			point.Departure = departures[index]
			index = index + 1
		}
		waypoints = append(waypoints, point)
	}
	assignment.WayPoints = waypoints
	return &assignment, nil
}

func createWaypointAssignment(coordinates [][2]float64, start Time) *Assignment {
	assignment := Assignment{Departure: start}
	waypoints := make([]WayPoint, 0, len(coordinates))
	for _, coordinate := range coordinates {
		waypoint := WayPoint{
			Name:      "custom waypoint",
			Latitude:  coordinate[0],
			Longitude: coordinate[1],
		}
		waypoints = append(waypoints, waypoint)
	}
	assignment.Name = "custom waypoint assignment"
	assignment.WayPoints = waypoints
	return &assignment
}
