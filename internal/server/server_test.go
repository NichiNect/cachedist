package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NichiNect/cachedist/internal/cache"
)

func TestServer_GetSet(t *testing.T) {
	c := cache.NewShardedCache(4, 100, 30)
	defer c.Stop()
	srv := NewServer(c, nil, nil)

	// Set value
	setBody := `{"key": "test_key", "value": "test_value", "ttl": 0}`
	reqSet := httptest.NewRequest(http.MethodPost, "/set", bytes.NewBufferString(setBody))
	reqSet.Header.Set("Content-Type", "application/json")
	wSet := httptest.NewRecorder()
	srv.ServeHTTP(wSet, reqSet)

	if wSet.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK for SET, got %v", wSet.Code)
	}

	// Get value
	reqGet := httptest.NewRequest(http.MethodGet, "/get?key=test_key", nil)
	wGet := httptest.NewRecorder()
	srv.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK for GET, got %v", wGet.Code)
	}

	var resp Response
	err := json.NewDecoder(wGet.Body).Decode(&resp)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if !resp.OK {
		t.Errorf("Expected resp.OK to be true")
	}

	if resp.Data.(string) != "test_value" {
		t.Errorf("Expected 'test_value', got %v", resp.Data)
	}
}
