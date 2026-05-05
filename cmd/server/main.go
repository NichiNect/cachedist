package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NichiNect/cachedist/config"
	"github.com/NichiNect/cachedist/internal/cache"
	"github.com/NichiNect/cachedist/internal/server"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting cachedist node %s...", cfg.NodeID)

	c := cache.NewSimpleCache()
	srv := server.NewServer(c)

	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: srv,
	}

	go func() {
		log.Printf("HTTP server listening on port %s", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown setup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
