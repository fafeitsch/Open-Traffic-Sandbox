package pax

import (
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	routing "github.com/fafeitsch/simple-timetable-routing"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConvertModel(t *testing.T) {
	mdl, _ := model.Init("../model/testdata/wuerzburg(fictional)")
	stops := convertModel(mdl.Lines())
	tt := routing.NewTimetable(stops.values())
	mainfrankenTheater := model.StopId("node/248513451")
	vogelVerlag := model.StopId("node/600918135")
	startTime, _ := time.Parse(time.Kitchen, "6:16AM")
	connection := tt.Query(stops[mainfrankenTheater], stops[vogelVerlag], startTime)
	assert.Equal(t, 2, len(connection.Legs), "number of legs")
}
