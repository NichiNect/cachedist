package server

import (
	"net/http"

	"github.com/NichiNect/cachedist/internal/cache"
)

// Server represents the HTTP server for the cache.
type Server struct {
	mux *http.ServeMux
}

// NewServer creates a new HTTP Server instance.
func NewServer(c cache.Cache) *Server {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/get", handleGet(c))
	mux.HandleFunc("/set", handleSet(c))
	mux.HandleFunc("/delete", handleDelete(c))
	mux.HandleFunc("/stats", handleStats(c))
	mux.HandleFunc("/keys", handleKeys(c))
	mux.HandleFunc("/health", handleHealth())
	
	return &Server{
		mux: mux,
	}
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
