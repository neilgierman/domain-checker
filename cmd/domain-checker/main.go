package main

// Main entrypoint. Everything is in other go files in the same package
func main() {
	connectDb()
	go queueProcessor()
	handleRequests()
	defer mongoReadClient.Disconnect(ctx)
	defer mongoWriteClient.Disconnect(ctx)
}
