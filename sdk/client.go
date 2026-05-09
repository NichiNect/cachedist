package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/NichiNect/cachedist/internal/cluster"
)

type Client struct {
	ring        *cluster.HashRing
	httpClients map[string]*http.Client
}

func NewClient(nodeAddresses []string) *Client {
	ring := cluster.NewHashRing()
	httpClients := make(map[string]*http.Client)

	for _, addr := range nodeAddresses {
		// Convert HTTP address to GRPC address to match the server's HashRing NodeID
		// This ensures the SDK and all Servers have the exact same HashRing mapping
		grpcAddr := strings.Replace(addr, "700", "800", 1)
		grpcAddr = strings.Replace(grpcAddr, "localhost", "127.0.0.1", 1)
		
		ring.AddNode(grpcAddr, addr) // using grpcAddr as NodeID, but storing HTTP addr for the client to hit
		httpClients[addr] = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	return &Client{
		ring:        ring,
		httpClients: httpClients,
	}
}

type setRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

type apiResponse struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

func (c *Client) Set(key, value string, ttl int) error {
	nodes := c.ring.GetNodes(key, 2)
	if len(nodes) == 0 {
		return fmt.Errorf("no nodes available")
	}

	var lastErr error
	for _, nodeAddr := range nodes {
		client := c.httpClients[nodeAddr]
		url := fmt.Sprintf("http://%s/set", nodeAddr)

		reqBody := setRequest{Key: key, Value: value, TTL: ttl}
		reqData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqData))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var apiResp apiResponse
			json.NewDecoder(resp.Body).Decode(&apiResp)
			lastErr = fmt.Errorf("failed to set key, status: %d, error: %s", resp.StatusCode, apiResp.Error)
			continue
		}

		return nil
	}

	return fmt.Errorf("set failed on all replicas. last error: %w", lastErr)
}

func (c *Client) Get(key string) (string, bool, error) {
	nodes := c.ring.GetNodes(key, 2)
	if len(nodes) == 0 {
		return "", false, fmt.Errorf("no nodes available")
	}

	var lastErr error
	for _, nodeAddr := range nodes {
		client := c.httpClients[nodeAddr]
		url := fmt.Sprintf("http://%s/get?key=%s", nodeAddr, key)

		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return "", false, nil
		}

		if resp.StatusCode != http.StatusOK {
			var apiResp apiResponse
			json.NewDecoder(resp.Body).Decode(&apiResp)
			lastErr = fmt.Errorf("failed to get key, status: %d, error: %s", resp.StatusCode, apiResp.Error)
			continue
		}

		var apiResp apiResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			lastErr = err
			continue
		}

		if !apiResp.OK {
			lastErr = fmt.Errorf("api error: %s", apiResp.Error)
			continue
		}

		if valStr, ok := apiResp.Data.(string); ok {
			return valStr, true, nil
		}
		
		lastErr = fmt.Errorf("unexpected data format")
	}

	return "", false, fmt.Errorf("get failed on all replicas. last error: %w", lastErr)
}

func (c *Client) Delete(key string) error {
	nodes := c.ring.GetNodes(key, 2)
	if len(nodes) == 0 {
		return fmt.Errorf("no nodes available")
	}

	var lastErr error
	for _, nodeAddr := range nodes {
		client := c.httpClients[nodeAddr]
		url := fmt.Sprintf("http://%s/delete?key=%s", nodeAddr, key)

		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return nil // idempotent delete
		}

		if resp.StatusCode != http.StatusOK {
			var apiResp apiResponse
			json.NewDecoder(resp.Body).Decode(&apiResp)
			lastErr = fmt.Errorf("failed to delete key, status: %d, error: %s", resp.StatusCode, apiResp.Error)
			continue
		}
		
		return nil
	}
	
	return fmt.Errorf("delete failed on all replicas. last error: %w", lastErr)
}
