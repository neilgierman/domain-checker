// This file contains REST endpoint implementations
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type DomainStatus int

const (
	Unknown DomainStatus = iota
	CatchAll
	NotCatchAll
)

func (d DomainStatus) String() string {
	return [...]string{"unknown", "catch-all", "not catch-all"}[d]
}

type DomainResult struct {
	DomainName string
	Status string
	DeliveredCount int
	BouncedCount int
}

// Everything that is not specifically implemented should just return a 404
func (a *App) defaultHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Path {} not found", r.URL.EscapedPath())
	w.WriteHeader(http.StatusNotFound)
}

// When a delivered request comes in, add it to the queue and send back a 202
func (a *App) queueDelivered(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedDomain := vars["domain"]
	entry := &queueEntry{Action: "delivered", Domain: requestedDomain}
	conn := a.connectRemoteQueue()
	defer conn.Close()
	ch := a.connectQueueChannel(conn)
	defer ch.Close()
	q := a.declareQueue(ch)
	body, err := json.Marshal(entry)
	if err != nil {
		log.Fatal(err)
	}
	a.publishMessage(&q, ch, body)
	w.WriteHeader(http.StatusAccepted)
}

// When a bounced request comes in, add it to the queue and send back a 202
func (a *App) queueBounced(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedDomain := vars["domain"]
	entry := &queueEntry{Action: "bounced", Domain: requestedDomain}
	conn := a.connectRemoteQueue()
	defer conn.Close()
	ch := a.connectQueueChannel(conn)
	defer ch.Close()
	q := a.declareQueue(ch)
	body, err := json.Marshal(entry)
	if err != nil {
		log.Fatal(err)
	}
	a.publishMessage(&q, ch, body)
	w.WriteHeader(http.StatusAccepted)
}

// When a get request comes in, process it immediately and send back the result
func (a *App) processGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domain := vars["domain"]
	domainEntry := a.getDomain(domain)
	domainResult := DomainResult{
		DomainName: domain,
		BouncedCount: domainEntry.BouncedCount,
		DeliveredCount: domainEntry.DeliveredCount,
	}
	if domainResult.DeliveredCount < 1000 {
		domainResult.Status = Unknown.String()
	} else if domainResult.BouncedCount == 0 {
		domainResult.Status = CatchAll.String()
	} else {
		domainResult.Status = NotCatchAll.String()
	}
	body, err := json.Marshal(domainResult)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type","application/json")
	w.Write(body)
}

