package server

import (
	"net/http"

	"github.com/NichiNect/cachedist/internal/cache"
	"github.com/NichiNect/cachedist/internal/cluster"
	"github.com/NichiNect/cachedist/internal/replication"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the HTTP server for the cache.
type Server struct {
	mux *http.ServeMux
}

// NewServer creates a new HTTP Server instance.
func NewServer(c cache.Cache, rep *replication.Replicator, mgr *cluster.Manager) *Server {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/get", instrumentedHandler("get", handleGet(c)))
	mux.HandleFunc("/set", instrumentedHandler("set", handleSet(c, rep)))
	mux.HandleFunc("/delete", instrumentedHandler("delete", handleDelete(c, rep)))
	mux.HandleFunc("/stats", instrumentedHandler("stats", handleStats(c)))
	mux.HandleFunc("/keys", instrumentedHandler("keys", handleKeys(c)))
	mux.HandleFunc("/health", instrumentedHandler("health", handleHealth()))
	
	mux.Handle("/metrics", promhttp.Handler())
	
	if mgr != nil {
		mux.HandleFunc("/cluster/join", instrumentedHandler("cluster_join", handleClusterJoin(mgr)))
	}
	
	return &Server{
		mux: mux,
	}
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
