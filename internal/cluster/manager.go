package cluster

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Manager struct {
	nodes    map[string]*NodeInfo
	ring     *HashRing
	mu       sync.RWMutex
	selfAddr string
	selfID   string
	syncer   Syncer
}

// Syncer interface to decouple from the actual sync implementation
type Syncer interface {
	SyncFromPeer(peerAddr string) error
}

func NewManager(ring *HashRing, selfID, selfAddr string) *Manager {
	return &Manager{
		nodes:    make(map[string]*NodeInfo),
		ring:     ring,
		selfID:   selfID,
		selfAddr: selfAddr,
	}
}

func (m *Manager) SetSyncer(s Syncer) {
	m.syncer = s
}

func (m *Manager) RegisterNode(info NodeInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If it's ourselves, ignore
	if info.ID == m.selfID {
		return
	}

	info.Status = StatusAlive
	info.LastSeen = time.Now()
	info.FailCount = 0

	m.nodes[info.ID] = &info
	m.ring.AddNode(info.ID, info.GRPCAddr)
	log.Printf("Node %s registered: %v", info.ID, info)
}

func (m *Manager) GetLiveNodes() []NodeInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var live []NodeInfo
	for _, n := range m.nodes {
		if n.Status == StatusAlive {
			live = append(live, *n)
		}
	}
	return live
}

func (m *Manager) StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	client := &http.Client{Timeout: 2 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.pingNodes(client)
		}
	}
}

func (m *Manager) pingNodes(client *http.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, node := range m.nodes {
		url := fmt.Sprintf("http://%s/health", node.HTTPAddr)
		resp, err := client.Get(url)
		
		success := err == nil && resp.StatusCode == http.StatusOK
		if err == nil {
			resp.Body.Close()
		}

		if success {
			node.FailCount = 0
			node.LastSeen = time.Now()

			if node.Status == StatusDead {
				log.Printf("Node %s is back online. Status -> Recovering", id)
				node.Status = StatusRecovering
				
				// Re-add to ring
				m.ring.AddNode(node.ID, node.GRPCAddr)

				// Trigger sync in background if syncer is set
				if m.syncer != nil {
					go func(n *NodeInfo) {
						log.Printf("Starting sync from peer %s...", n.ID)
						err := m.syncer.SyncFromPeer(n.HTTPAddr)
						if err != nil {
							log.Printf("Sync from %s failed: %v", n.ID, err)
						} else {
							log.Printf("Sync from %s completed successfully", n.ID)
							
							m.mu.Lock()
							n.Status = StatusAlive
							m.mu.Unlock()
						}
					}(node)
				} else {
					// If no syncer, just mark alive
					node.Status = StatusAlive
				}
			}
		} else {
			node.FailCount++
			if node.FailCount >= 3 && node.Status == StatusAlive {
				log.Printf("Node %s failed 3 consecutive heartbeats. Marking as DEAD.", id)
				node.Status = StatusDead
				m.ring.RemoveNode(node.ID)
			}
		}
	}
}
