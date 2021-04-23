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
	a.Initialize("administrator","5VhVTxQb272kMsm","cluster0.k6tho.mongodb.net","domain-checker")
	go a.queueProcessor()
	a.handleRequests()
	a.Run()
	defer a.ReadClient.Disconnect(ctx)
	defer a.WriteClient.Disconnect(ctx)
}
