package main

import (
	"container/list"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type App struct {
	Router *mux.Router
	ReadClient *mongo.Client
	WriteClient *mongo.Client
	ReadDb *mongo.Database
	WriteDb *mongo.Database
	Queue *list.List
}

func (a *App) Initialize(user, password, host, database string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// We are creating a different client for reads and writes
	// This is done in case we need the application to connect to different
	//  infrastructure for example to read from a local cache and write to
	//  a more central instance
	var err error
	a.WriteClient, err = mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb+srv://" + user + ":" + password + "@" + host + "/" + database + "?retryWrites=true&w=majority",
	))
	if err != nil { 
		log.Fatal(err)
	}
	err = a.WriteClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	a.WriteDb = a.WriteClient.Database(database)
	
	a.ReadClient, err = mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb+srv://" + user + ":" + password + "@" + host + "/" + database + "?retryWrites=true&w=majority",
	))
	if err != nil { 
		log.Fatal(err)
	}
	err = a.ReadClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	a.ReadDb = a.ReadClient.Database(database)
	log.Print("DB Connected")

	a.Queue = list.New()
	a.handleRequests()
}

// Sets up the handlers for the different REST endpoints
func (a *App) handleRequests() {
	a.Router = mux.NewRouter().StrictSlash(true)
	a.Router.HandleFunc("/events/{domain}/delivered", a.queueDelivered).Methods("PUT")
	a.Router.HandleFunc("/events/{domain}/bounced", a.queueBounced).Methods("PUT")
	a.Router.HandleFunc("/domains/{domain}", a.processGet).Methods("GET")
	a.Router.PathPrefix("/").HandlerFunc(a.defaultHandler)
}


func (a *App) Run() {
	log.Print("Listening on Port 80")
	http.ListenAndServe(":80", a.Router)
}
