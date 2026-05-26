package main

import (
	"log"
	"net/http"
	"time"

	"rewrite/internal/api"
	"rewrite/internal/config"
)

func main() {
	cfg := config.Load()

	srv := &http.Server{
		Addr:         cfg.HttpRootAddress,
		Handler:      api.NewHandler(&cfg),
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server starting at %s", srv.Addr)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
