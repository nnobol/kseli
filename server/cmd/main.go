package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"kseli-server/config"
	"kseli-server/handlers"
	"kseli-server/handlers/middleware"
	"kseli-server/services"
	"kseli-server/storage"
)

func main() {
	config.LoadConfig()

	st := storage.InitializeStorage()
	rs := services.NewRoomService(st)

	mux := http.NewServeMux()

	appDir := "../builds/client"
	fileServer := http.FileServer(http.Dir(appDir))

	// POST request to create a chat room
	mux.Handle("POST /api/rooms", middleware.WithMiddleware(
		handlers.CreateRoomHandler(rs),
		middleware.ValidateFingerprint(),
		middleware.ValidateSessionID(),
		middleware.ValidateAPIKey(),
		middleware.ValidateOrigin(),
	))

	// POST request for a user to join a chat room
	mux.Handle("POST /api/rooms/{roomID}/join", middleware.WithMiddleware(
		handlers.JoinRoomHandler(rs),
		middleware.ValidateFingerprint(),
		middleware.ValidateSessionID(),
		middleware.ValidateAPIKey(),
		middleware.ValidateOrigin(),
	))

	// GET request to get chat room details
	mux.Handle("GET /api/rooms/{roomID}", middleware.WithMiddleware(
		handlers.GetRoomHandler(rs),
		middleware.ValidateAuthToken(),
		middleware.ValidateOrigin(),
	))

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

	log.Println("Listening on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
