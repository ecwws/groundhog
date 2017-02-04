package main

import (
	"time"
)

func worker(termination <-chan bool, notify chan<- bool) {

MainLoop:
	for {
		select {
		case <-termination:
			Logger.Debug.Println("Worker received termination request")
			break MainLoop
		default:
			time.Sleep(2 * time.Second)
		}
	}

	Logger.Debug.Println("Notifying main thread on worker termination")
	notify <- true
}
