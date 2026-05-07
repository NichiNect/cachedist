package cluster

import (
	"fmt"
	"hash/fnv"
	"sort"
	"sync"
)

const CACHE_VIRTUAL_NODES = 150

type HashRing struct {
	nodes  map[uint32]string // maps virtual node hash to physical node address
	sorted []uint32          // sorted slice of virtual node hashes
	mu     sync.RWMutex
}

func NewHashRing() *HashRing {
	return &HashRing{
		nodes:  make(map[uint32]string),
		sorted: make([]uint32, 0),
	}
}

func fnv32(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (r *HashRing) AddNode(nodeID, address string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < CACHE_VIRTUAL_NODES; i++ {
		hash := fnv32(fmt.Sprintf("%s#%d", nodeID, i))
		r.nodes[hash] = address
	}

	r.rebuildSorted()
}

func (r *HashRing) RemoveNode(nodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < CACHE_VIRTUAL_NODES; i++ {
		hash := fnv32(fmt.Sprintf("%s#%d", nodeID, i))
		delete(r.nodes, hash)
	}

	r.rebuildSorted()
}

func (r *HashRing) rebuildSorted() {
	var newSorted []uint32
	for hash := range r.nodes {
		newSorted = append(newSorted, hash)
	}

	sort.Slice(newSorted, func(i, j int) bool {
		return newSorted[i] < newSorted[j]
	})
	r.sorted = newSorted
}

func (r *HashRing) GetNode(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.sorted) == 0 {
		return ""
	}

	hash := fnv32(key)

	idx := sort.Search(len(r.sorted), func(i int) bool {
		return r.sorted[i] >= hash
	})

	if idx == len(r.sorted) {
		idx = 0
	}

	return r.nodes[r.sorted[idx]]
}
