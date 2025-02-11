package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	// POST request to create a chat room, auth using API key
	mux.Handle("POST /api/rooms", middleware.WithMiddleware(
		handlers.CreateRoomHandler(storage),
		middleware.ValidateAPIKey(),
		middleware.ValidateOrigin(),
	))

	// GET request to get chat room details, using Auth token
	mux.Handle("GET /api/rooms/{roomID}", middleware.WithMiddleware(
		handlers.GetRoomHandler(storage),
		middleware.ValidateAuthToken(),
		middleware.ValidateOrigin(),
	))

	// POST request for a user to join a chat room, auth using API key
	mux.Handle("POST /api/rooms/{roomID}/users", middleware.WithMiddleware(
		handlers.JoinRoomHandler(storage),
		middleware.ValidateAPIKey(),
		middleware.ValidateOrigin(),
	))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			handlers.WriteJSONError(w, http.StatusNotFound, "Resource Not Found.", nil)
			return
		}

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

	log.Println("Listening on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
