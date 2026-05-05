package server

import (
	"encoding/json"
	"net/http"

	"github.com/NichiNect/cachedist/internal/cache"
)

// Response represents the standard API response format.
type Response struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

func sendJSON(w http.ResponseWriter, status int, data interface{}, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		OK:    errMsg == "",
		Data:  data,
		Error: errMsg,
	})
}

// handleGet handles GET /get?key={key}
func handleGet(c cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendJSON(w, http.StatusMethodNotAllowed, nil, "method not allowed")
			return
		}
		
		key := r.URL.Query().Get("key")
		if key == "" {
			sendJSON(w, http.StatusBadRequest, nil, "key is required")
			return
		}

		val, found := c.Get(key)
		if !found {
			sendJSON(w, http.StatusNotFound, nil, "key not found")
			return
		}

		sendJSON(w, http.StatusOK, val, "")
	}
}

type setRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

// handleSet handles POST /set
func handleSet(c cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendJSON(w, http.StatusMethodNotAllowed, nil, "method not allowed")
			return
		}

		var req setRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendJSON(w, http.StatusBadRequest, nil, "invalid request body: "+err.Error())
			return
		}

		if req.Key == "" {
			sendJSON(w, http.StatusBadRequest, nil, "key is required")
			return
		}

		err := c.Set(req.Key, req.Value, req.TTL)
		if err != nil {
			sendJSON(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		sendJSON(w, http.StatusOK, "OK", "")
	}
}

// handleDelete handles DELETE /delete?key={key}
func handleDelete(c cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			sendJSON(w, http.StatusMethodNotAllowed, nil, "method not allowed")
			return
		}

		key := r.URL.Query().Get("key")
		if key == "" {
			sendJSON(w, http.StatusBadRequest, nil, "key is required")
			return
		}

		c.Delete(key)
		sendJSON(w, http.StatusOK, "OK", "")
	}
}

// handleStats handles GET /stats
func handleStats(c cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendJSON(w, http.StatusMethodNotAllowed, nil, "method not allowed")
			return
		}

		stats := c.Stats()
		sendJSON(w, http.StatusOK, stats, "")
	}
}

// handleKeys handles GET /keys
func handleKeys(c cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendJSON(w, http.StatusMethodNotAllowed, nil, "method not allowed")
			return
		}

		keys := c.Keys()
		sendJSON(w, http.StatusOK, keys, "")
	}
}

// handleHealth handles GET /health
func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendJSON(w, http.StatusMethodNotAllowed, nil, "method not allowed")
			return
		}
		sendJSON(w, http.StatusOK, "OK", "")
	}
}
