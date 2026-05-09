package cluster

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/NichiNect/cachedist/internal/cache"
)

type NodeSyncer struct {
	cache    cache.Cache
	ring     *HashRing
	selfAddr string
}

func NewNodeSyncer(c cache.Cache, ring *HashRing, selfAddr string) *NodeSyncer {
	return &NodeSyncer{
		cache:    c,
		ring:     ring,
		selfAddr: selfAddr,
	}
}

// syncResponse represents the expected response from GET /keys
type syncResponse struct {
	OK   bool `json:"ok"`
	Data []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		TTL   int    `json:"ttl"`
	} `json:"data"`
	Error string `json:"error"`
}

func (s *NodeSyncer) SyncFromPeer(peerAddr string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("http://%s/keys", peerAddr)

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to GET /keys: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResp syncResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("api error: %s", apiResp.Error)
	}

	for _, item := range apiResp.Data {
		// Verify if we are responsible for this key
		nodes := s.ring.GetNodes(item.Key, 2)
		isResponsible := false
		for _, n := range nodes {
			if n == s.selfAddr {
				isResponsible = true
				break
			}
		}
		
		if isResponsible {
			s.cache.Set(item.Key, item.Value, item.TTL)
		}
	}
	return nil
}
