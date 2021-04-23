// This file contains REST endpoint implementations
package main

import (
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Everything that is not specifically implemented should just return a 404
func (a *App) defaultHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Path {} not found", r.URL.EscapedPath())
	w.WriteHeader(http.StatusNotFound)
}

// When a delivered request comes in, add it to the queue and send back a 202
func (a *App) queueDelivered(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedDomain := vars["domain"]
	entry := queueEntry{action: "delivered", domain: requestedDomain}
	queue.PushBack(entry)
	w.WriteHeader(http.StatusAccepted)
}

// When a bounced request comes in, add it to the queue and send back a 202
func (a *App) queueBounced(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedDomain := vars["domain"]
	entry := queueEntry{action: "bounced", domain: requestedDomain}
	queue.PushBack(entry)
	w.WriteHeader(http.StatusAccepted)
}

// When a get request comes in, process it immediately and send back the result
func (a *App) processGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domain := vars["domain"]
	domainResult := a.getDomain(domain)
	if domainResult.DeliveredCount < 1000 {
		io.WriteString(w, "unknown")
	} else if domainResult.BouncedCount == 0 {
		io.WriteString(w, "catch-all")
	} else {
		io.WriteString(w, "not catch-all")
	}
}

