package rest

import (
	"encoding/json"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"github.com/gorilla/mux"
	"net/http"
)

type busInfo struct {
	Id         model.BusId `json:"id"`
	Assignment string      `json:"assignment"`
	Line       *restLine   `json:"line,omitempty"`
}

func (a *api) getBusInfo(w http.ResponseWriter, r *http.Request) {
	bus, ok := a.findBus(w, r)
	if !ok {
		return
	}
	assignment := a.dispatcher.QueryCurrentAssignment(bus.Id)
	result := busInfo{
		Id:         bus.Id,
		Assignment: assignment.Name,
	}
	if assignment.Line != nil {
		line := mapToRestLine(*assignment.Line)
		result.Line = &line
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(result)
}

func (a *api) findBus(w http.ResponseWriter, r *http.Request) (*model.Bus, bool) {
	id, ok := mux.Vars(r)["key"]
	bus, ok := a.busModel.Bus(model.BusId(id))
	if !ok {
		errorResponse(w, http.StatusNotFound, "could not find bus with id \"%s\"", id)
		return &model.Bus{}, false
	}
	return bus, true
}

func (a *api) getRouteOfBus(w http.ResponseWriter, r *http.Request) {
	bus, ok := a.findBus(w, r)
	if !ok {
		return
	}
	assignment := a.dispatcher.QueryCurrentAssignment(bus.Id)
	coords := make([]model.Coordinate, 0, len(assignment.WayPoints))
	for _, wp := range assignment.WayPoints {
		coords = append(coords, wp)
	}
	a.queryAndWriteRouteToWriter(w, coords)
}

func (a *api) queryAndWriteRouteToWriter(w http.ResponseWriter, coords []model.Coordinate) {
	route, _, err := a.gps(coords...)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not query routes: %v", err)
		return
	}
	result := make([][2]float64, 0, len(route))
	for _, coordinate := range route {
		result = append(result, [2]float64{coordinate.Lat(), coordinate.Lon()})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(result)
}
