package model

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInit(t *testing.T) {
	mdl, err := Init("./testdata/wuerzburg(fictional)")
	require.NoError(t, err, "no error expected")
	require.Equal(t, 5, len(mdl.Buses()), "number of buses in the scenario.")

	bus2, _ := mdl.Bus(BusId("V2"))
	assert.Equal(t, BusId("V2"), bus2.Id, "Id of the first bus")
	assert.Equal(t, "", bus2.Name, "Name of the first bus")
	require.Equal(t, 2, len(bus2.Assignments), "number of assignments of the first bus")
	assignment := bus2.Assignments[1]
	assert.Equal(t, "Busbahnhof - Residenz - Sanderau", assignment.Name, "Name of assignment")
	line := assignment.Line
	assert.Equal(t, 11, len(assignment.WayPoints), "number of waypoints in the line")
	id := StopId("node/248513451")
	assert.Equal(t, WayPoint{Departure: 24000000, Id: &id, Name: "Mainfranken Theater", Latitude: 49.7947734, Longitude: 9.9360743}, assignment.WayPoints[3], "sample waypoint")

	assert.Equal(t, assignment.Name, line.Name, "name of the line")
	assert.Equal(t, 11, len(line.waypoints), "number of stops in the line")
	assert.Equal(t, 39, len(line.departures[*line.waypoints[8].Id]), "number of tours in the line")

	assignment = bus2.Assignments[0]
	assert.Equal(t, "custom waypoint assignment", assignment.Name, "name of assignment")
	assert.Equal(t, 2, len(assignment.WayPoints), "number of waypoints")
	assert.Equal(t, WayPoint{Departure: 0, Id: nil, Name: "custom waypoint", Latitude: 49.8012835, Longitude: 9.9340999}, assignment.WayPoints[1], "second waypoint")
}
