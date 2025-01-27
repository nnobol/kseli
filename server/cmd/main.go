package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"kseli-server/config"
	"kseli-server/handlers"
	"kseli-server/middleware"
	"kseli-server/storage"
)

func main() {
	config.LoadConfig()

	storage := storage.NewMemoryStore()
	mux := http.NewServeMux()

	appDir := "../builds/client"
	fileServer := http.FileServer(http.Dir(appDir))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestRoute := filepath.Join(appDir, r.URL.Path)

		_, err := os.Stat(requestRoute)
		if os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(appDir, "index.html"))
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fileServer.ServeHTTP(w, r)
	}))

	// POST request to create a chat room, public, using API key
	mux.Handle("/api/room", middleware.WithMiddleware(
		handlers.CreateRoomHandler(storage),
		middleware.ValidateAPIKey(),
		middleware.ValidateOrigin(),
		middleware.ValidateHTTPMethod(http.MethodPost),
	))

	log.Println("Listening on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
