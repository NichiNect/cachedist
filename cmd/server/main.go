package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/NichiNect/cachedist/config"
	"github.com/NichiNect/cachedist/internal/cache"
	"github.com/NichiNect/cachedist/internal/cluster"
	"github.com/NichiNect/cachedist/internal/grpc"
	"github.com/NichiNect/cachedist/internal/replication"
	"github.com/NichiNect/cachedist/internal/server"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting cachedist node %s...", cfg.NodeID)

	c := cache.NewShardedCache(cfg.NumShards, cfg.MaxItems, cfg.TTLCleanup)
	defer c.Stop()
	
	ring := cluster.NewHashRing()
	selfAddr := fmt.Sprintf("127.0.0.1:%s", cfg.GRPCPort) // assuming localhost for peers
	ring.AddNode(cfg.NodeID, selfAddr)

	if cfg.Peers != "" {
		peers := strings.Split(cfg.Peers, ",")
		for i, peer := range peers {
			// A simple way to generate node IDs for peers, e.g., peer-0, peer-1
			ring.AddNode(fmt.Sprintf("peer-%d", i), peer)
		}
	}

	rep := replication.NewReplicator(ring, selfAddr)

	go func() {
		grpcserver.StartServer(":"+cfg.GRPCPort, c)
	}()

	srv := server.NewServer(c, rep)

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
