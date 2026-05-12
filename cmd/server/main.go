package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/NichiNect/cachedist/internal/metrics"
	"github.com/NichiNect/cachedist/internal/replication"
	"github.com/NichiNect/cachedist/internal/server"
)

func main() {
	cfg := config.Load()
	metrics.NodeID = cfg.NodeID

	log.Printf("Starting cachedist node %s...", cfg.NodeID)

	c := cache.NewShardedCache(cfg.NumShards, cfg.MaxItems, cfg.TTLCleanup)
	defer c.Stop()
	
	ring := cluster.NewHashRing()
	
	// Determine the host to advertise. Use NodeID if it's not the default "node-1"
	// (which usually means we are in Docker or a custom setup)
	advertiseHost := "127.0.0.1"
	if cfg.NodeID != "node-1" {
		advertiseHost = cfg.NodeID
	}
	
	selfAddr := fmt.Sprintf("%s:%s", advertiseHost, cfg.GRPCPort)
	ring.AddNode(selfAddr, selfAddr) 

	if cfg.Peers != "" {
		peers := strings.Split(cfg.Peers, ",")
		for _, peer := range peers {
			ring.AddNode(peer, peer)
		}
	}

	rep := replication.NewReplicator(ring, selfAddr)

	mgr := cluster.NewManager(ring, cfg.NodeID, selfAddr)
	syncer := cluster.NewNodeSyncer(c, ring, selfAddr)
	mgr.SetSyncer(syncer)

	// Register peers in the manager
	if cfg.Peers != "" {
		peers := strings.Split(cfg.Peers, ",")
		
		info := cluster.NodeInfo{
			ID:       cfg.NodeID,
			HTTPAddr: fmt.Sprintf("%s:%s", advertiseHost, cfg.HTTPPort),
			GRPCAddr: selfAddr,
		}
		infoData, _ := json.Marshal(info)

		for i, peer := range peers {
			peerID := fmt.Sprintf("peer-%d", i)
			
			peerParts := strings.Split(peer, ":")
			peerHost := peerParts[0]
			peerGRPCPort := peerParts[1]
			
			// For local testing, we assume peer HTTP port is peer GRPC port - 1000
			peerHTTPPort := strings.Replace(peerGRPCPort, "8", "7", 1)
			peerHTTPAddr := fmt.Sprintf("%s:%s", peerHost, peerHTTPPort)
			
			mgr.RegisterNode(cluster.NodeInfo{
				ID:       peerID,
				HTTPAddr: peerHTTPAddr,
				GRPCAddr: peer,
			})
			
			// Announce ourselves to peers asynchronously
			go func(pAddr string) {
				url := fmt.Sprintf("http://%s/cluster/join", pAddr)
				http.Post(url, "application/json", bytes.NewBuffer(infoData))
			}(peerHTTPAddr)
		}
	}

	go func() {
		grpcserver.StartServer(":"+cfg.GRPCPort, c)
	}()

	srv := server.NewServer(c, rep, mgr)

	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: srv,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mgr.StartHeartbeat(ctx)

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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
