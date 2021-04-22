// This file contains the queue implementation
package main

import (
	"container/list"
	"log"
	"time"
)

// Global queue variable holds the write requests
var queue = list.New()

// This function is expected to be launched in its own go routine (thread)
// It will monitor the queue and process queue entries as they arrive
func queueProcessor() {
	log.Print("Starting queue processor")
	for {
		// If the queue is empty wait a second
		// No reason to chew up CPU cycles on a queue that is empty
		if queue.Len() == 0 {
			time.Sleep(1 * time.Second)
			continue
		}
		item := queue.Front()
		processPut(queueEntry(item.Value.(queueEntry)))
		queue.Remove(item)
	}
}

