// This file has all of the back end database functionality in it
package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// Get a single domain record from the database
func (a *App) getDomain(domain string) domainEntry {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var domainResult domainEntry
	domainsCollection := a.ReadClient.Database("domain-checker").Collection("domains")
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
	var newEntry = false
	domainEntry := a.getDomain(item.domain)
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
	default:
		// We shouldn't really get here but if we do, don't do anything with a DB record
		log.Print("action {} not implemented", item.action)
		return
	}
	if !newEntry {
		update := bson.M{
			"$set": domainEntry,
		}
		result, err := a.WriteClient.Database("domain-checker").Collection("domains").UpdateOne(
			ctx,
			bson.M{"_id": domainEntry.ID},
			update,
		)
		if err != nil {
			log.Print(err)
		}
		log.Print(result)
	} else {
		result, err := a.WriteClient.Database("domain-checker").Collection("domains").InsertOne(
			ctx,
			domainEntry,
		)
		if err != nil {
			log.Print(err)
		}
		log.Print(result)
		
	}
}