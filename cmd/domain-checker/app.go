package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type App struct {
	router *mux.Router
	dbConfig *DBConfig
	appCfg *Config
}

type DBConfig struct {
	readClient *mongo.Client
	writeClient *mongo.Client
	readDb *mongo.Database
	writeDb *mongo.Database
	writeDbLock *sync.Mutex
}

func (a *App) Initialize(host, port, database string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// We are creating a different client for reads and writes
	// This is done in case we need the application to connect to different
	//  infrastructure for example to read from a local cache and write to
	//  a more central instance
	var err error
	a.dbConfig.writeClient, err = mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb://" + host + ":" + port,
	))
	if err != nil { 
		return err
	}
	err = a.dbConfig.writeClient.Ping(ctx, readpref.Primary())
	if err != nil {
		return err
	}
	a.dbConfig.writeDb = a.dbConfig.writeClient.Database(database)
	
	a.dbConfig.readClient, err = mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb://" + host + ":" + port,
	))
	if err != nil { 
		return err
	}
	err = a.dbConfig.readClient.Ping(ctx, readpref.Primary())
	if err != nil {
		return err
	}
	a.dbConfig.readDb = a.dbConfig.readClient.Database(database)
	log.Print("DB Connected")

	a.handleRequests()
	return nil
}

// Sets up the handlers for the different REST endpoints
func (a *App) handleRequests() {
	a.router = mux.NewRouter().StrictSlash(true)
	a.router.HandleFunc("/events/{domain}/delivered", a.queueDelivered).Methods("PUT")
	a.router.HandleFunc("/events/{domain}/bounced", a.queueBounced).Methods("PUT")
	a.router.HandleFunc("/domains/{domain}", a.processGet).Methods("GET")
	a.router.PathPrefix("/").HandlerFunc(a.defaultHandler)
}


func (a *App) Run() {
	log.Print("Listening on Port 80")
	http.ListenAndServe(":80", a.router)
}
