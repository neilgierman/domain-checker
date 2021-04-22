// This file has all of the back end database functionality in it
package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// This is a structure for the format of an update entry in the queue
type queueEntry struct {
	action string
	domain string
}

// This is the structure of the document in the MongoDB
type domainEntry struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	DomainName string `bson:"domainName,omitempty"`
	DeliveredCount int `bson:"deliveredCount,omitempty"`
	BouncedCount int `bson:"bouncedCount,omitempty"`
}

// We are creating a different client for reads and writes
// This is done in case we need the application to connect to different
//  infrastructure for example to read from a local cache and write to
//  a more central instance
var mongoReadDb *mongo.Database
var mongoReadClient *mongo.Client
var mongoWriteDb *mongo.Database
var mongoWriteClient *mongo.Client
var ctx context.Context

func connectDb() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoWriteClient, err := mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb+srv://administrator:5VhVTxQb272kMsm@cluster0.k6tho.mongodb.net/domain-checker?retryWrites=true&w=majority",
	))
	if err != nil { 
		log.Fatal(err)
	}
	err = mongoWriteClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	mongoWriteDb = mongoWriteClient.Database("domain-checker")

	mongoReadClient, err := mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb+srv://administrator:5VhVTxQb272kMsm@cluster0.k6tho.mongodb.net/domain-checker?retryWrites=true&w=majority",
	))
	if err != nil { 
		log.Fatal(err)
	}
	err = mongoReadClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	mongoReadDb = mongoReadClient.Database("domain-checker")
	log.Print("DB Connected")
}

func getDomain(domain string) domainEntry {
	var domainResult domainEntry
	domainsCollection := mongoReadDb.Collection("domains")
	// If we got an error, it could be because there isn't a document with that domain created yet.
	// Set up a new domainEntry with a specific count for delivered and bounced so the DB code
	//  knows if this is an insert new instead of update existing.
	if err := domainsCollection.FindOne(ctx, bson.M{"domainName": domain}).Decode(&domainResult); err != nil {
		log.Print(err)
		return domainEntry{
			DomainName: domain,
			DeliveredCount: -1,
			BouncedCount: -1,
		}
	}
	log.Print(domainResult)
	return domainResult
}

// Do the update/insert of a count increase for a domain
func processPut(item queueEntry) {
	log.Print(item)
	var newEntry = false
	domainEntry := getDomain(item.domain)
	// If the counts are -1, then the domain didn't already exist in the database
	// Set a flag so we know this is a new record and not an update to an
	//  existing record
	if domainEntry.DeliveredCount == -1 && domainEntry.BouncedCount == -1 {
		domainEntry.DeliveredCount = 0
		domainEntry.BouncedCount = 0
		newEntry = true
	}
	switch item.action {
	case "bounced":
		domainEntry.BouncedCount++
	case "delivered":
		domainEntry.DeliveredCount++
	}
	if !newEntry {
		update := bson.M{
			"$set": domainEntry,
		}
		result, err := mongoWriteDb.Collection("domains").UpdateOne(
			ctx,
			bson.M{"_id": domainEntry.ID},
			update,
		)
		if err != nil {
			log.Print(err)
		}
		log.Print(result)
	} else {
		result, err := mongoWriteDb.Collection("domains").InsertOne(
			ctx,
			domainEntry,
		)
		if err != nil {
			log.Print(err)
		}
		log.Print(result)
		
	}
}