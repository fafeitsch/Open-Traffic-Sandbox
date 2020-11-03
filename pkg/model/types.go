package model

import "time"

type BusId string

type Bus struct {
	Id          BusId
	Name        string
	Assignments []Assignment
}

type Assignment struct {
	Name      string
	Departure Time
	WayPoints []WayPoint
}

type WayPoint struct {
	IsStop    bool
	Name      string
	Latitude  float64
	Longitude float64
}

func (w *WayPoint) Lat() float64 {
	return w.Latitude
}

func (w *WayPoint) Lon() float64 {
	return w.Longitude
}

type Coordinate interface {
	Lat() float64
	Lon() float64
}

// RouteService is an interface capable of computing detailed waypoints between the provided waypoints.
type RouteService func(Coordinate, Coordinate) ([]Coordinate, float64, error)

type BusPosition struct {
	BusId    BusId      `json:"id"`
	Location [2]float64 `json:"loc"`
}

type Publisher func(position BusPosition)

type Time int

func (t Time) HourMinute() (int, int) {
	minutes := int(t) / 1000 / 60 / 60
	return minutes / 24, minutes - minutes/24
}

type Timer struct {
	HeartBeat     <-chan Time
	originalTimer time.Timer
}

func (t *Timer) Stop() bool {
	return t.originalTimer.Stop()
}

func NewTimer(interval time.Duration, start Time) Timer {
	last := time.Now()
	originalTimer := time.NewTimer(interval)
	channel := make(chan Time)
	go func() {
		current := start
		for t := range originalTimer.C {
			gap := t.Sub(last)
			current = Time(int(current) + int(gap*time.Millisecond))
			t = last
			channel <- current
		}
	}()
	return Timer{HeartBeat: channel}
}
