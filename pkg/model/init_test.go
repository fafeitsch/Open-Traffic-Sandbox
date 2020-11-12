package model

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInit(t *testing.T) {
	mdl, err := Init("./test-data/wuerzburg(fictional)")
	require.NoError(t, err, "no error expected")
	require.Equal(t, 2, len(mdl.Buses()), "number of buses in the scenario.")

	bus2 := mdl.Buses()[1]
	assert.Equal(t, BusId("V2"), bus2.Id, "Id of the first bus")
	assert.Equal(t, "", bus2.Name, "Name of the first bus")
	require.Equal(t, 2, len(bus2.Assignments), "number of assignments of the first bus")
	assignment := bus2.Assignments[1]
	assert.Equal(t, "Busbahnhof - Residenz - Sanderau", assignment.Name, "Name of assignment")
	line := assignment.Line
	assert.Equal(t, 9, len(assignment.WayPoints), "number of waypoints in the line")
	assert.Equal(t, WayPoint{Departure: 24000000, IsStop: true, Name: "Mainfranken Theater", Latitude: 49.7947186, Longitude: 9.9359725}, assignment.WayPoints[2], "sample waypoint")

	assert.Equal(t, assignment.Name, line.Name, "name of the line")
	assert.Equal(t, 9, len(line.Stops), "number of stops in the line")
	assert.Equal(t, 39, len(line.departures[line.Stops[8].id]), "number of tours in the line")

	assignment = bus2.Assignments[0]
	assert.Equal(t, "", assignment.Name, "name of assignment")
	assert.Equal(t, 2, len(assignment.WayPoints), "number of waypoints")
	assert.Equal(t, WayPoint{Departure: 0, IsStop: false, Name: "custom waypoint", Latitude: 49.8012835, Longitude: 9.9340999}, assignment.WayPoints[1], "second waypoint")
}