package channels

import "github.com/fafeitsch/Open-Traffic-Sandbox/routing"

// Merge takes a slice of channels and returns one single channel
// which emits data put into any of the input channels.
func Merge(channels []<-chan routing.VehicleLocation) <-chan routing.VehicleLocation {
	if len(channels) == 0 {
		return nil
	}
	if len(channels) == 1 {
		return channels[0]
	}
	m := len(channels) / 2
	channel1 := Merge(channels[:m])
	channel2 := Merge(channels[m:])
	return mergeTwoChannels(channel1, channel2)
}

func mergeTwoChannels(a, b <-chan routing.VehicleLocation) <-chan routing.VehicleLocation {
	c := make(chan routing.VehicleLocation)
	go func() {
		defer close(c)
		for a != nil || b != nil {
			select {
			case v, ok := <-a:
				if !ok {
					a = nil
					continue
				}
				c <- v
			case v, ok := <-b:
				if !ok {
					b = nil
					continue
				}
				c <- v
			}
		}
	}()
	return c
}
