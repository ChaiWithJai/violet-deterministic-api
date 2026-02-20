package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/restarone/violet-deterministic-api/internal/config"
	httpapi "github.com/restarone/violet-deterministic-api/internal/http"
)

func main() {
	cfg := config.Load()
	srv, err := httpapi.NewServer(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}
	log.Printf("violet-deterministic-api listening on :%s", cfg.Port)

	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("server exited: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
