package server

import (
	"net/http"

	"github.com/NichiNect/cachedist/internal/cache"
	"github.com/NichiNect/cachedist/internal/cluster"
	"github.com/NichiNect/cachedist/internal/replication"
)

// Server represents the HTTP server for the cache.
type Server struct {
	mux *http.ServeMux
}

// NewServer creates a new HTTP Server instance.
func NewServer(c cache.Cache, rep *replication.Replicator, mgr *cluster.Manager) *Server {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/get", handleGet(c))
	mux.HandleFunc("/set", handleSet(c, rep))
	mux.HandleFunc("/delete", handleDelete(c, rep))
	mux.HandleFunc("/stats", handleStats(c))
	mux.HandleFunc("/keys", handleKeys(c))
	mux.HandleFunc("/health", handleHealth())
	
	if mgr != nil {
		mux.HandleFunc("/cluster/join", handleClusterJoin(mgr))
	}
	
	return &Server{
		mux: mux,
	}
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
