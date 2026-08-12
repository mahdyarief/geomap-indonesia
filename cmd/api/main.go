package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/mahdyarief/geomap-indonesia/internal/config"
	"github.com/mahdyarief/geomap-indonesia/internal/database"
	"github.com/mahdyarief/geomap-indonesia/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	r := router.New(cfg, pool)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("geomap-indonesia API listening on :%s (%s)", cfg.ServerPort, cfg.AppEnv)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
