package main

import (
	"fmt"
	fileMonitor "lxraa/file_monitor"
)

func ReadGlobalConfig() {

}

func main() {
	channel := make(chan interface{})
	fileMonitor.Start(channel)
	defer fileMonitor.End()
	for {
		event := <-channel
		switch event := event.(type) {
		case *fileMonitor.FileEvent:
			for i := 0; i < len(event.Events); i++ {
				fmt.Println(event.Events[i].EventDec)
			}
		}
	}
}
