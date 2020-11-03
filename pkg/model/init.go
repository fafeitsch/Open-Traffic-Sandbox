package model

import (
	routing "github.com/fafeitsch/simple-timetable-routing"
)

type BusModel interface {
	Buses() []Bus
}

type Model interface {
	BusModel
}

func Init(directory string) Model {
	return nil
}

type scenario struct {
	Start routing.Time
	Buses []ioBus
}

type location struct {
	Lat       float64
	Lon       float64
	Reference string
}

type ioBus struct {
	Id          string
	Assignments []ioAssignment
}

type ioAssignment struct {
	Start routing.Time
	Line  *string
	GoTo  *location `yaml:"goTo"`
}

type model struct {
	scenario *scenario
}
