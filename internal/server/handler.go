package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/NichiNect/cachedist/internal/cache"
	"github.com/NichiNect/cachedist/internal/cluster"
	"github.com/NichiNect/cachedist/internal/replication"
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
func handleSet(c cache.Cache, rep *replication.Replicator) http.HandlerFunc {
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

		if rep != nil {
			replicatedTo, err := rep.Replicate(req.Key, req.Value, req.TTL, false)
			if err != nil {
				sendJSON(w, http.StatusInternalServerError, nil, "replication failed: "+err.Error())
				return
			}
			if len(replicatedTo) > 0 {
				w.Header().Set("X-Replicated-To", strings.Join(replicatedTo, ","))
			}
		}

		sendJSON(w, http.StatusOK, "OK", "")
	}
}

// handleDelete handles DELETE /delete?key={key}
func handleDelete(c cache.Cache, rep *replication.Replicator) http.HandlerFunc {
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

		if rep != nil {
			replicatedTo, err := rep.Replicate(key, "", 0, true)
			if err != nil {
				sendJSON(w, http.StatusInternalServerError, nil, "replication failed: "+err.Error())
				return
			}
			if len(replicatedTo) > 0 {
				w.Header().Set("X-Replicated-To", strings.Join(replicatedTo, ","))
			}
		}

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

		items := c.GetAllItems()
		
		type itemResp struct {
			Key   string `json:"key"`
			Value string `json:"value"`
			TTL   int    `json:"ttl"`
		}
		
		var responseData []itemResp
		for key, item := range items {
			var ttl int
			if !item.ExpiresAt.IsZero() {
				// Calculate remaining TTL
				ttl = int(time.Until(item.ExpiresAt).Seconds())
				if ttl < 0 {
					continue // expired
				}
			}
			responseData = append(responseData, itemResp{
				Key:   key,
				Value: string(item.Value),
				TTL:   ttl,
			})
		}

		sendJSON(w, http.StatusOK, responseData, "")
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

// handleClusterJoin handles POST /cluster/join
func handleClusterJoin(mgr *cluster.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendJSON(w, http.StatusMethodNotAllowed, nil, "method not allowed")
			return
		}

		var info cluster.NodeInfo
		if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
			sendJSON(w, http.StatusBadRequest, nil, "invalid body")
			return
		}

		mgr.RegisterNode(info)
		sendJSON(w, http.StatusOK, "Registered", "")
	}
}
