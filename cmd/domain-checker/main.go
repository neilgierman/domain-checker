package main

import (
	"context"
	"log"
	"time"
)

// Main entrypoint. Everything is in other go files in the same package
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	a := App{}
	a.loadConfig()
	err := a.Initialize(a.appCfg.Database.Host, a.appCfg.Database.Port, a.appCfg.Database.Database)
	if err != nil {
		log.Fatal(err)
	}
	go a.queueProcessor()
	go a.databaseBatchWriter()
	a.handleRequests()
	a.Run()
	defer a.dbConfig.readClient.Disconnect(ctx)
	defer a.dbConfig.writeClient.Disconnect(ctx)
}
