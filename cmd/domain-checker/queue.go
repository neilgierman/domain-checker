// This file contains the queue implementation
package main

import (
	"log"
	"time"
)

// This function is expected to be launched in its own go routine (thread)
// It will monitor the queue and process queue entries as they arrive
func (a *App) queueProcessor() {
	log.Print("Starting queue processor")
	for {
		// If the queue is empty wait a second
		// No reason to chew up CPU cycles on a queue that is empty
		if a.Queue.Len() == 0 {
			time.Sleep(1 * time.Second)
			continue
		}
		item := a.Queue.Front()
		a.processPut(queueEntry(item.Value.(queueEntry)))
		a.Queue.Remove(item)
	}
}

