package model

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func createDepartures(csv []string) ([]Time, error) {
	firstTime, err := ParseTime(csv[2])
	if err != nil {
		return nil, fmt.Errorf("third column must be in time format hh:mm, but was \"%s\"", csv[2])
	}
	result := make([]Time, 0, len(csv))
	result = append(result, firstTime)
	intervalRegex := regexp.MustCompile("every (\\d+) min")
	var currentInterval *time.Duration
	for index, departureTime := range csv[3:] {
		match := intervalRegex.FindStringSubmatch(departureTime)
		if match != nil {
			minutes, _ := strconv.Atoi(match[1])
			interval := time.Duration(minutes) * time.Minute
			currentInterval = &interval
			if index+4 > len(csv)-1 {
				return nil, fmt.Errorf("an interval column must be followed by an absolute time colum")
			}
			nextAbsoluteTime, err := ParseTime(csv[index+4])
			if err != nil {
				return nil, fmt.Errorf("column %d is a interval column, but is not succeeded by a valid absolute time column", index+3)
			}
			nextTime := result[len(result)-1].Add(*currentInterval)
			for ok := true; ok; ok = nextTime.Before(nextAbsoluteTime) {
				result = append(result, nextTime)
				nextTime = result[len(result)-1].Add(*currentInterval)
			}
		} else {
			parsed, err := ParseTime(departureTime)
			if err != nil {
				return nil, fmt.Errorf("column %d with content \"%s\" is neither a valid time column nor an interval column", index+3, departureTime)
			}
			result = append(result, parsed)
		}
	}
	return result, nil
}
