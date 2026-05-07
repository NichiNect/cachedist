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
		ring.AddNode(addr, addr) // using address as nodeID
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
	nodeAddr := c.ring.GetNode(key)
	if nodeAddr == "" {
		return fmt.Errorf("no nodes available")
	}

	client := c.httpClients[nodeAddr]
	url := fmt.Sprintf("http://%s/set", nodeAddr)

	reqBody := setRequest{Key: key, Value: value, TTL: ttl}
	reqData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiResp apiResponse
		json.NewDecoder(resp.Body).Decode(&apiResp)
		return fmt.Errorf("failed to set key, status: %d, error: %s", resp.StatusCode, apiResp.Error)
	}

	return nil
}

func (c *Client) Get(key string) (string, bool, error) {
	nodeAddr := c.ring.GetNode(key)
	if nodeAddr == "" {
		return "", false, fmt.Errorf("no nodes available")
	}

	client := c.httpClients[nodeAddr]
	url := fmt.Sprintf("http://%s/get?key=%s", nodeAddr, key)

	resp, err := client.Get(url)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", false, nil
	}

	if resp.StatusCode != http.StatusOK {
		var apiResp apiResponse
		json.NewDecoder(resp.Body).Decode(&apiResp)
		return "", false, fmt.Errorf("failed to get key, status: %d, error: %s", resp.StatusCode, apiResp.Error)
	}

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", false, err
	}

	if !apiResp.OK {
		return "", false, fmt.Errorf("api error: %s", apiResp.Error)
	}

	// Assuming the value is a string based on server implementation
	if valStr, ok := apiResp.Data.(string); ok {
		return valStr, true, nil
	}
	
	return "", false, fmt.Errorf("unexpected data format")
}

func (c *Client) Delete(key string) error {
	nodeAddr := c.ring.GetNode(key)
	if nodeAddr == "" {
		return fmt.Errorf("no nodes available")
	}

	client := c.httpClients[nodeAddr]
	url := fmt.Sprintf("http://%s/delete?key=%s", nodeAddr, key)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil // idempotent delete
	}

	if resp.StatusCode != http.StatusOK {
		var apiResp apiResponse
		json.NewDecoder(resp.Body).Decode(&apiResp)
		return fmt.Errorf("failed to delete key, status: %d, error: %s", resp.StatusCode, apiResp.Error)
	}
	return nil
}
