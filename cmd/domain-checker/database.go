// This file has all of the back end database functionality in it
package main

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

var batchUpdates []mongo.WriteModel

// Get a single domain record from the database
func (a *App) getDomain(domain string) domainEntry {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var domainResult domainEntry
	domainsCollection := a.dbConfig.readDb.Collection("domains")
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
	log.Print(item)

	batchUpdate := mongo.NewUpdateOneModel()
	batchUpdate.SetFilter(bson.M{"domainName":item.Domain})
	batchUpdate.SetUpsert(true)

	switch item.Action {
	case "bounced":
		batchUpdate.SetUpdate(bson.M{
			"$inc": bson.M{
				"bouncedCount":1,
			},
		})
	case "delivered":
		batchUpdate.SetUpdate(bson.M{
			"$inc": bson.M{
				"deliveredCount":1,
			},
		})
	default:
		// We shouldn't really get here but if we do, don't do anything with a DB record
		log.Print("action {} not implemented", item.Action)
		return
	}

	a.dbConfig.writeDbLock.Lock()
	batchUpdates = append(batchUpdates, batchUpdate)
	log.Print(batchUpdate, " added to batchUpdates")
	log.Print("batchUpdates is ", len(batchUpdates))
	a.dbConfig.writeDbLock.Unlock()
}

// This function is expected to be launched in its own go routine (thread)
// It will monitor the queue and process queue entries as they arrive
func (a *App) databaseBatchWriter() {
	log.Print("Starting DB queue processor")

	a.dbConfig.writeDbLock = &sync.Mutex{}

	bulkOption := options.BulkWriteOptions{}
	bulkOption.SetOrdered(true)

	for {
		select {
		case <- time.After(2 * time.Second):
			if len(batchUpdates) > 0 {
				a.dbConfig.writeDbLock.Lock()
				log.Print("Starting batch for ", len(batchUpdates), " records")
				log.Print(batchUpdates)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				result, err := a.dbConfig.writeDb.Collection("domains").BulkWrite(ctx, batchUpdates, &bulkOption)
				if err != nil {
					log.Print("BulkWrite error: ",err)
				}
				log.Print("result: ", result)
				batchUpdates = nil
				a.dbConfig.writeDbLock.Unlock()
			}
		}
	}

}
