package model

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

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

var TimeRegex = regexp.MustCompile("^([0-9]+):?([0-5][0-9])$")

func ParseTime(time string) (Time, error) {
	submatch := TimeRegex.FindStringSubmatch(string(time))
	if submatch == nil {
		return 0, fmt.Errorf("the string \"%s\" does not match the required format", time)
	}
	hour, _ := strconv.Atoi(submatch[1])
	minute, _ := strconv.Atoi(submatch[2])
	return Time((hour*60 + minute) * 60 * 1000), nil
}

func (t Time) HourMinute() (int, int) {
	minutes := int(t) / 1000 / 60
	hours := minutes / 60
	return hours, minutes - hours*60
}

func (t Time) Before(other Time) bool {
	return t < other
}

func (t Time) Sub(other Time) time.Duration {
	return time.Duration(t-other) * time.Millisecond
}

func (t Time) String() string {
	hour, minute := t.HourMinute()
	return fmt.Sprintf("%02d:%02d", hour, minute)
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

type StopId string

type Stop struct {
	id        StopId
	name      string
	latitude  float64
	longitude float64
}

func (s Stop) Lat() float64 {
	return s.latitude
}

func (s Stop) Lon() float64 {
	return s.longitude
}

func (s Stop) String() string {
	return fmt.Sprintf("%s(%s)", s.name, s.id)
}
