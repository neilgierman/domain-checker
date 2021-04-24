// This file has all of the back end database functionality in it
package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// This is a structure for the format of an update entry in the queue
type queueEntry struct {
	Action string `json:"action"`
	Domain string `json:"domain"`
}

// This is the structure of the document in the MongoDB
type domainEntry struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	DomainName string `bson:"domainName,omitempty"`
	DeliveredCount int `bson:"deliveredCount,omitempty"`
	BouncedCount int `bson:"bouncedCount,omitempty"`
}

// Get a single domain record from the database
func (a *App) getDomain(domain string) domainEntry {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var domainResult domainEntry
	domainsCollection := a.ReadDb.Collection("domains")
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
func (a *App) processPut(item queueEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Print(item)

	filter := bson.M{"domainName":item.Domain}

	var update primitive.M
	switch item.Action {
	case "bounced":
		update = bson.M{
			"$inc": bson.M{
				"bouncedCount":1,
			},
		}
	case "delivered":
		update = bson.M{
			"$inc": bson.M{
				"deliveredCount":1,
			},
		}
	default:
		// We shouldn't really get here but if we do, don't do anything with a DB record
		log.Print("action {} not implemented", item.Action)
		return
	}

	upsert := true
	after:= options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert: &upsert,
	}

	result := a.WriteDb.Collection("domains").FindOneAndUpdate(ctx, filter, update, &opt)

	log.Print(result)
}