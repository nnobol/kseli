package main

import (
	"log"
	"net/http"
	"time"

	"kseli/config"
	"kseli/router"
)

func main() {
	config.LoadConfig()
	mux := router.New()

	srv := &http.Server{
		Addr:              ":3000",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    1 << 20,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	log.Println("Listening on :3000...")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
