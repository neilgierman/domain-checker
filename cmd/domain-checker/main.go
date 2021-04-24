package main

import (
	"context"
	"time"
)

// Main entrypoint. Everything is in other go files in the same package
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	a := App{}
	a.loadConfig()
	a.Initialize(a.appCfg.Database.Host, a.appCfg.Database.Port, a.appCfg.Database.Database)
	go a.queueProcessor()
	a.handleRequests()
	a.Run()
	defer a.dbConfig.readClient.Disconnect(ctx)
	defer a.dbConfig.writeClient.Disconnect(ctx)
}
