package model

import (
	"fmt"
	"time"
)

func Example_newTicker() {
	start, _ := ParseTime("7:34")
	ticker := NewTicker(1*time.Second, start)
	time1 := <-ticker.HeartBeat
	time2 := <-ticker.HeartBeat
	time3 := <-ticker.HeartBeat
	ticker.Stop()
	_ = <-ticker.HeartBeat
	fmt.Printf("%d – %d - %d", time1, time2, time3)
	// Output:
}
