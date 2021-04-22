// This file contains REST endpoint implementations
package main

import (
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Sets up the handlers for the different REST endpoints
func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/events/{domain}/delivered", queueDelivered).Methods("PUT")
	myRouter.HandleFunc("/events/{domain}/bounced", queueBounced).Methods("PUT")
	myRouter.HandleFunc("/domains/{domain}", processGet).Methods("GET")
	myRouter.PathPrefix("/").HandlerFunc(defaultHandler)
	log.Print("Listening on Port 80")
	http.ListenAndServe(":80", myRouter)

}

// Everything that is not specifically implemented should just return a 404
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Path {} not found", r.URL.EscapedPath())
	w.WriteHeader(http.StatusNotFound)
}

// When a delivered request comes in, add it to the queue and send back a 202
func queueDelivered(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedDomain := vars["domain"]
	entry := queueEntry{action: "delivered", domain: requestedDomain}
	queue.PushBack(entry)
	w.WriteHeader(http.StatusAccepted)
}

// When a bounced request comes in, add it to the queue and send back a 202
func queueBounced(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedDomain := vars["domain"]
	entry := queueEntry{action: "bounced", domain: requestedDomain}
	queue.PushBack(entry)
	w.WriteHeader(http.StatusAccepted)
}

// When a get request comes in, process it immediately and send back the result
func processGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domain := vars["domain"]
	domainResult := getDomain(domain)
	if domainResult.DeliveredCount < 1000 {
		io.WriteString(w, "unknown")
	} else if domainResult.BouncedCount == 0 {
		io.WriteString(w, "catch-all")
	} else {
		io.WriteString(w, "not catch-all")
	}
}

