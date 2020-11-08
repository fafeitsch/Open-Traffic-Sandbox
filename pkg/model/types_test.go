package model

import (
	"fmt"
	"time"
)

func ExampleNewTicker() {
	start, _ := ParseTime("7:34")
	ticker := NewTicker(start, 1*time.Minute, 1*time.Millisecond)
	time1 := <-ticker.HeartBeat
	time2 := <-ticker.HeartBeat
	time3 := <-ticker.HeartBeat
	ticker.Stop()
	_ = <-ticker.HeartBeat
	fmt.Printf("%v - %v - %v", time1, time2, time3)
	// Output: 07:34 - 07:35 - 07:36
}

func ExampleParseTime() {
	moment, _ := ParseTime("15:04")
	fmt.Printf("Milliseconds since midnight: %d\n", moment)
	hour, minute := moment.HourMinute()
	fmt.Printf("Hour: %d, Minute: %d\n", hour, minute)
	later := moment.Add(86 * time.Minute)
	fmt.Printf("%v is later than %v: %v (difference: %d minutes)", later, moment, moment.Before(later), later.Sub(moment)/time.Minute)
	// Output: Milliseconds since midnight: 54240000
	// Hour: 15, Minute: 4
	// 16:30 is later than 15:04: true (difference: 86 minutes)
}

func ExampleLine_TourTimes() {
	stops := []*Stop{&Stop{id: "stopA"}, &Stop{id: "stopB"}, &Stop{id: "stopC"}, &Stop{id: "stopD"}}
	baseTime, _ := ParseTime("16:35")
	departures := map[StopId][]Time{
		"stopA": {baseTime, baseTime.Add(7 * time.Minute), baseTime.Add(14 * time.Minute)},
		"stopB": {baseTime.Add(2 * time.Minute), baseTime.Add(9 * time.Minute), baseTime.Add(16 * time.Minute)},
		"stopC": {baseTime.Add(3 * time.Minute), baseTime.Add(12 * time.Minute), baseTime.Add(19 * time.Minute)},
		"stopD": {baseTime.Add(1 * time.Minute), baseTime.Add(13 * time.Minute), baseTime.Add(20 * time.Minute)},
	}
	line := Line{Stops: stops, departures: departures}
	fmt.Printf("Departures for second tour: %v\n", line.TourTimes(baseTime.Add(7*time.Minute)))
	fmt.Printf("Returns nil if tour not found: %v\n", line.TourTimes(baseTime.Add(2*time.Minute)) == nil)
	// Output: Departures for second tour: [16:42 16:44 16:47 16:48]
	// Returns nil if tour not found: true
}
