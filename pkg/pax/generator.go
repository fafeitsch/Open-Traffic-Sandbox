package pax

import (
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	routing "github.com/fafeitsch/simple-timetable-routing"
)

type Generator struct {
	Lines []model.Line
}

func (g *Generator) Start(frequency float64, warp float64) {

}

func convertModel(lines []model.Line) stopMapping {
	stops, routeLines := createRoutingStops(lines)
	for _, line := range lines {
		startTimes := line.StartTimes()
		lineStops := line.Stops()
		for _, time := range startTimes {
			departures := line.TourTimes(time)
			for index, departure := range departures[0 : len(departures)-1] {
				travelTime := departures[index+1].Sub(departure)
				stop := lineStops[index+1]
				event := routing.Event{Line: routeLines[line.Id], Departure: routing.CreateTime(departure.HourMinute()), NextStop: stops[*stop.Id], TravelTime: travelTime}
				currentStop := lineStops[index]
				stops[*currentStop.Id].Events = append(stops[*currentStop.Id].Events, event)
			}
		}
	}
	return stops
}

func createRoutingStops(lines []model.Line) (map[model.StopId]*routing.Stop, map[model.LineId]*routing.Line) {
	resStops := make(map[model.StopId]*routing.Stop)
	resLines := make(map[model.LineId]*routing.Line)
	for _, line := range lines {
		for _, stop := range line.Stops() {
			if stop.Id != nil {
				resStops[*stop.Id] = routing.NewStop(string(*stop.Id), stop.Name)
			}
		}
		resLines[line.Id] = &routing.Line{Name: line.Name, Id: string(line.Id)}
	}
	return resStops, resLines
}

type stopMapping map[model.StopId]*routing.Stop

func (s stopMapping) values() []*routing.Stop {
	result := make([]*routing.Stop, 0, len(s))
	for _, stop := range s {
		result = append(result, stop)
	}
	return result
}
