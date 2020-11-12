package model

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// BusId is used to identify a Bus.
type BusId string

// Bus describes a vehicle can can load passengers.
type Bus struct {
	Id          BusId
	Name        string
	Assignments []Assignment
}

// Assignment is a task for a Bus to do.
type Assignment struct {
	Name      string
	Line      *Line
	Departure Time
	WayPoints []WayPoint
}

// WayPoint is a part of an assignment.
type WayPoint struct {
	Departure Time
	IsStop    bool
	Name      string
	Latitude  float64
	Longitude float64
}

func (w WayPoint) Lat() float64 {
	return w.Latitude
}

func (w WayPoint) Lon() float64 {
	return w.Longitude
}

// Coordinate is an interface providing latitude and longitude data.
type Coordinate interface {
	Lat() float64
	Lon() float64
}

// RouteService is a function capable of computing detailed waypoints between the provided waypoints.
type RouteService func(Coordinate, Coordinate) ([]Coordinate, float64, error)

// BusPosition describes the position of a certain bus at the current moment. BusPosition is meant
// to be sent to subscribers, possible over network. Thus, we keep this struct small.
type BusPosition struct {
	BusId    BusId      `json:"id"`
	Location [2]float64 `json:"loc"`
}

// Publisher is a function taking care to broadcast BusPosition updates.
type Publisher func(position BusPosition)

// Time specifies the time of the day in milliseconds. The difference to time.Time is
// that Time does not specify the date. A Time can be parsed from a kitchen clock string such as "15:04"
// with ParseTime.
type Time int

var timeRegex = regexp.MustCompile("^([0-9]+):?([0-5][0-9])$")

// ParseTime creates a new Time from a string. The time string must be a time given in time.Kitchen 24hour format,
// such as "15:04". If the time string is not parsable, then an error is returned. In this case, the returned time is 0.
func ParseTime(timeString string) (Time, error) {
	subMatch := timeRegex.FindStringSubmatch(string(timeString))
	if subMatch == nil {
		return 0, fmt.Errorf("the string \"%s\" does not match the required format", timeString)
	}
	hour, _ := strconv.Atoi(subMatch[1])
	minute, _ := strconv.Atoi(subMatch[2])
	return Time((hour*60 + minute) * 60 * 1000), nil
}

// MustParseTime haves nearly identical to ParseTime. The only difference is that MustParseTime will panic
// if the timeString cannot be parsed, whereas ParseTime returns a non-nil error.
func MustParseTime(timeString string) Time {
	result, err := ParseTime(timeString)
	if err != nil {
		panic(err)
	}
	return result
}

// HourMinute returns the hour and the minute of the time.
func (t Time) HourMinute() (int, int) {
	minutes := int(t) / 1000 / 60
	hours := minutes / 60
	return hours, minutes - hours*60
}

// Before returns true if this time is strictly before the other time.
func (t Time) Before(other Time) bool {
	return t < other
}

// Add creates a new Time which represents this time plus the provided duration.
func (t Time) Add(duration time.Duration) Time {
	return Time(int(t) + int(duration/time.Millisecond))
}

// Sub computes the difference between this time and the given time. It is negative if
// this time is before (see Time.Before) the other time.
func (t Time) Sub(other Time) time.Duration {
	return time.Duration(t-other) * time.Millisecond
}

func (t Time) String() string {
	hour, minute := t.HourMinute()
	return fmt.Sprintf("%02d:%02d", hour, minute)
}

// Ticker is similar to time.Ticker as it contains a channel that produces
// a Time in specified intervals. The difference to time.Ticker is that this Ticker
// has a specified start time and operates on Time, rather than on time.Time. New tickers
// should be created with NewTicker.
type Ticker struct {
	HeartBeat      <-chan Time
	heartBeat      chan Time
	originalTicker *time.Ticker
}

// Stop prevents the ticker from emitting more events and closes the writing channel.
func (t *Ticker) Stop() {
	t.originalTicker.Stop()
	close(t.heartBeat)
}

// NewTicker creates and starts a new ticker. The start parameter specifies the first Time to be emited
// by the ticker. The simulation interval defines which duration lies between two successive emits of the ticker.
// The real interval is the real duration between two successive emits. This third interval may not be accurate (see time.NewTicker).
func NewTicker(start Time, simulationInterval time.Duration, realInterval time.Duration) Ticker {
	originalTicker := time.NewTicker(realInterval)
	channel := make(chan Time)
	go func() {
		last := start
		for _ = range originalTicker.C {
			channel <- last
			last = last.Add(simulationInterval)
		}
	}()
	return Ticker{heartBeat: channel, HeartBeat: channel, originalTicker: originalTicker}
}

// StopId is used to identify a Stop.
type StopId string

// Stop represents a location where buses stop and let passengers enter and exit the bus.
type Stop struct {
	id        StopId
	name      string
	latitude  float64
	longitude float64
}

func (s *Stop) Lat() float64 {
	return s.latitude
}

func (s *Stop) Lon() float64 {
	return s.longitude
}

func (s *Stop) String() string {
	return fmt.Sprintf("%s(%s)", s.name, s.id)
}

// LineId is used to identify a Line.
type LineId string

// Line represents a predefined path and departures times for buses.
type Line struct {
	Id         LineId
	Name       string
	Stops      []*Stop
	departures map[StopId][]Time
}

func (l *Line) String() string {
	return fmt.Sprintf("%s(%s)", l.Name, l.Id)
}

// TourTimes returns all departure times of the tour starting at start.
// If no tour of this line starts at the given time, then nil is returned.
// If the line is not well defined (e.g. no Stops, no adequate departures) then the
// behaviour of this method is not well defined. It will most likely panic.
func (l *Line) TourTimes(start Time) []Time {
	departures := l.departures[l.Stops[0].id]
	index := 0
	departure := departures[0]
	for departure != start && index < len(departures)-1 {
		index = index + 1
		departure = departures[index]
	}
	if departure != start {
		return nil
	}
	result := make([]Time, 0, len(departures))
	for _, stop := range l.Stops {
		result = append(result, l.departures[stop.id][index])
	}
	return result
}
