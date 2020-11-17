package model

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

func loadLines(scenario scenario, directory string, stops map[StopId]Stop) ([]Line, error) {
	result := make([]Line, 0, len(scenario.Lines))
	for _, line := range scenario.Lines {
		loadedLine, err := loadLineFromFile(filepath.Join(directory, line.File), stops)
		if err != nil {
			return nil, fmt.Errorf("could not parse line \"%s\": %v", line.Id, err)
		}
		loadedLine.Id = LineId(line.Id)
		loadedLine.Name = line.Name
		result = append(result, *loadedLine)
	}
	return result, nil
}

func loadLineFromFile(filePath string, stops map[StopId]Stop) (*Line, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("loading line file failed: %v", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.ReuseRecord = true
	reader.LazyQuotes = true
	stopList := make([]*Stop, 0, 0)
	departureMap := make(map[StopId][]Time)
	for data, err := reader.Read(); err == nil; data, err = reader.Read() {
		if ok, wayPointOnly := isEntryWaypointOnly(data); ok {
			stopList = append(stopList, wayPointOnly)
			continue
		}
		stopId := StopId(data[1])
		stop, ok := stops[stopId]
		if !ok {
			return nil, fmt.Errorf("could not find stop \"%s\"", stopId)
		}
		stopList = append(stopList, &stop)
		departures, err := createDepartures(data)
		if err != nil {
			return nil, fmt.Errorf("could not parse departures: %v", err)
		}
		departureMap[stopId] = departures
	}
	return &Line{Stops: stopList, departures: departureMap}, nil
}

func createDepartures(csvLine []string) ([]Time, error) {
	firstTime, err := ParseTime(csvLine[2])
	if err != nil {
		return nil, fmt.Errorf("third column must be in time format hh:mm, but was \"%s\"", csvLine[2])
	}
	result := make([]Time, 0, len(csvLine))
	result = append(result, firstTime)
	intervalRegex := regexp.MustCompile("every (\\d+) min")
	var currentInterval *time.Duration
	for index, departureTime := range csvLine[3:] {
		match := intervalRegex.FindStringSubmatch(departureTime)
		if match != nil {
			minutes, _ := strconv.Atoi(match[1])
			interval := time.Duration(minutes) * time.Minute
			currentInterval = &interval
			if index+4 > len(csvLine)-1 {
				return nil, fmt.Errorf("an interval column must be followed by an absolute time colum")
			}
			nextAbsoluteTime, err := ParseTime(csvLine[index+4])
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

func isEntryWaypointOnly(csv []string) (bool, *Stop) {
	allExcept2ndColEmpty := csv[0] == ""
	for _, entry := range csv[2:] {
		allExcept2ndColEmpty = allExcept2ndColEmpty && entry == ""
	}
	if !allExcept2ndColEmpty {
		return false, nil
	}
	coordinateRegex := regexp.MustCompile("^([0-9]+(?:.[0-9]+));([0-9]+(?:.[0-9]+))$")
	subMatch := coordinateRegex.FindStringSubmatch(csv[1])
	if len(subMatch) < 3 {
		return false, nil
	}
	latitude, _ := strconv.ParseFloat(subMatch[1], 64)
	longitude, _ := strconv.ParseFloat(subMatch[2], 64)
	return true, &Stop{WayPoint: WayPoint{Latitude: latitude, Longitude: longitude, Name: "custom waypoint"}}
}
