package main

import (
	"log"
	"net/http"

	"kseli-server/config"
	"kseli-server/handlers"
	"kseli-server/middleware"
	"kseli-server/storage"
)

func main() {
	config.LoadConfig()

	storage := storage.NewMemoryStore()
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("../builds/client"))
	mux.Handle("/", fileServer)

	// POST request to create a chat room, public, using API key
	mux.Handle("/api/room", middleware.WithMiddleware(
		handlers.CreateRoomHandler(storage),
		middleware.ValidateAPIKey(),
		middleware.ValidateHTTPMethod(http.MethodPost),
	))

	log.Println("Listening on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
