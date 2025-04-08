package main

import (
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kseli/config"
	"kseli/features/chat"
	"kseli/middleware"
	"kseli/storage"
)

func main() {
	config.LoadConfig()

	storage := storage.InitializeStorage()

	mux := http.NewServeMux()

	env := os.Getenv("ENV")
	clientDir := "./client"
	if env == "local" {
		clientDir = "../builds/client"
	}

	// POST request to create a chat room
	mux.Handle("POST /api/rooms", middleware.WithMiddleware(
		chat.CreateRoomHandler(storage),
		middleware.ValidateParticipantSessionID(),
		middleware.ValidateAPIKey(),
		middleware.ValidateOrigin(),
	))

	// POST request for a participant to join a chat room
	mux.Handle("POST /api/rooms/{roomID}/join", middleware.WithMiddleware(
		chat.JoinRoomHandler(storage),
		middleware.ValidateParticipantSessionID(),
		middleware.ValidateAPIKey(),
		middleware.ValidateOrigin(),
	))

	// GET request to get chat room details
	mux.Handle("GET /api/rooms/{roomID}", middleware.WithMiddleware(
		chat.GetRoomHandler(storage),
		middleware.ValidateToken(),
		middleware.ValidateOrigin(),
	))

	// DELETE request to close the chat room
	mux.Handle("DELETE /api/rooms/{roomID}", middleware.WithMiddleware(
		chat.DeleteRoomHandler(storage),
		middleware.ValidateToken(),
		middleware.ValidateOrigin(),
	))

	// POST request to kick a participant from a chat room
	mux.Handle("POST /api/rooms/{roomID}/kick", middleware.WithMiddleware(
		chat.KickParticipantHandler(storage),
		middleware.ValidateToken(),
		middleware.ValidateOrigin(),
	))

	// POST request to ban a user from a chat room
	mux.Handle("POST /api/rooms/{roomID}/ban", middleware.WithMiddleware(
		chat.BanParticipantHandler(storage),
		middleware.ValidateToken(),
		middleware.ValidateOrigin(),
	))

	mux.Handle("/ws/room", chat.RoomWSHandler(storage))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(clientDir, r.URL.Path)
		info, err := os.Stat(filePath)
		if os.IsNotExist(err) || (err == nil && info.IsDir()) {
			// Fallback to index.html for non-existent files or directories
			http.ServeFile(w, r, filepath.Join(clientDir, "index.html"))
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		acceptEncoding := r.Header.Get("Accept-Encoding")

		if strings.Contains(acceptEncoding, "br") {
			brPath := filePath + ".br"
			if fileExists(brPath) {
				setContentHeaders(w, filePath, "br")
				http.ServeFile(w, r, brPath)
				return
			}
		}

		if strings.Contains(acceptEncoding, "gzip") {
			gzPath := filePath + ".gz"
			if fileExists(gzPath) {
				setContentHeaders(w, filePath, "gzip")
				http.ServeFile(w, r, gzPath)
				return
			}
		}

		setContentHeaders(w, filePath, "")
		http.ServeFile(w, r, filePath)
	}))

	srv := &http.Server{
		Addr:              ":3000",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	log.Println("Listening on :3000...")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func setContentHeaders(w http.ResponseWriter, originalPath string, encoding string) {
	// Set correct Content-Type based on the original file extension
	ext := filepath.Ext(originalPath)
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	}

	if encoding != "" {
		w.Header().Set("Content-Encoding", encoding)
		w.Header().Set("Vary", "Accept-Encoding")
	}

	// caching for hashed files
	if strings.Contains(filepath.Base(originalPath), ".") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
}
