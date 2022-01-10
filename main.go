package main

import (
	"fmt"
	"lxraa/monitor"
)

func main() {

	channel := make(chan interface{})
	monitor.StartFileM(channel)
	defer monitor.EndFileM()
	for {
		event := <-channel
		switch event := event.(type) {
		case *monitor.FileEvent:
			for i := 0; i < len(event.Events); i++ {
				fmt.Println(event.Events[i].EventDec)
			}
		}
	}
}
