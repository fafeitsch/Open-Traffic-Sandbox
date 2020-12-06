package rest

import (
	"encoding/json"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
)

type restStop struct {
	Name      string
	Id        model.StopId
	Latitude  float64
	Longitude float64
}

type restLine struct {
	Id    model.LineId `json:"id"`
	Name  string       `json:"name"`
	Color string       `json:"color"`
}

func (a *api) getLines(w http.ResponseWriter, r *http.Request) {
	lines := a.lineModel.Lines()
	sort.Slice(lines, func(i, j int) bool {
		return lines[i].DefinitionIndex < lines[j].DefinitionIndex
	})
	result := make([]restLine, 0, len(lines))
	for _, line := range lines {
		result = append(result, restLine{
			Id:    line.Id,
			Name:  line.Name,
			Color: line.Color,
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(result)
}

func (a *api) getLine(w http.ResponseWriter, r *http.Request) {
	line, ok := a.findLine(w, r)
	if !ok {
		return
	}
	result := mapToRestLine(line)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(result)
}

func mapToRestLine(line model.Line) restLine {
	result := restLine{
		Id:    line.Id,
		Name:  line.Name,
		Color: line.Color,
	}
	return result
}

func (a *api) getRoute(w http.ResponseWriter, r *http.Request) {
	line, ok := a.findLine(w, r)
	if !ok {
		return
	}
	coords := make([]model.Coordinate, 0, len(line.Stops))
	for _, stop := range line.Stops {
		coords = append(coords, stop)
	}
	a.queryAndWriteRouteToWriter(w, coords)
}

func (a *api) findLine(w http.ResponseWriter, r *http.Request) (model.Line, bool) {
	id, ok := mux.Vars(r)["key"]
	line, ok := a.lineModel.Line(model.LineId(id))
	if !ok {
		errorResponse(w, http.StatusNotFound, "could not find line with id \"%s\"", id)
		return model.Line{}, false
	}
	return line, true
}
