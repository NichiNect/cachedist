package replication

import (
	"log"
	"sync"
	"time"

	"github.com/NichiNect/cachedist/internal/cluster"
	"github.com/NichiNect/cachedist/internal/grpc"
	"github.com/NichiNect/cachedist/pkg/pb"
)

type Replicator struct {
	grpcClient *grpcserver.GRPCClient
	ring       *cluster.HashRing
	selfAddr   string
}

func NewReplicator(ring *cluster.HashRing, selfAddr string) *Replicator {
	return &Replicator{
		grpcClient: grpcserver.NewGRPCClient(),
		ring:       ring,
		selfAddr:   selfAddr,
	}
}

// Replicate sends the data to the 2 replicas next in the hash ring.
// Returns an error if quorum is not reached (0 acks).
func (r *Replicator) Replicate(key, value string, ttl int, isDelete bool) ([]string, error) {
	// Find nodes for the key
	nodes := r.ring.GetNodes(key, 3) // Get up to 3 nodes (primary + 2 replicas)
	
	var replicas []string
	for _, node := range nodes {
		if node != r.selfAddr {
			replicas = append(replicas, node)
		}
	}

	if len(replicas) == 0 {
		return nil, nil // No replicas, trivial success
	}

	operation := pb.Operation_SET
	if isDelete {
		operation = pb.Operation_DELETE
	}

	req := &pb.ReplicateRequest{
		Key:        key,
		Value:      value,
		TtlSeconds: int64(ttl),
		Operation:  operation,
	}

	var wg sync.WaitGroup
	ackCount := 0
	var mu sync.Mutex
	replicatedTo := make([]string, 0)

	// Timeout for the entire replication process
	done := make(chan struct{})

	for _, replica := range replicas {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			err := r.grpcClient.Replicate(addr, req)
			if err != nil {
				log.Printf("Replication to %s failed: %v", addr, err)
			} else {
				mu.Lock()
				ackCount++
				replicatedTo = append(replicatedTo, addr)
				mu.Unlock()
			}
		}(replica)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		log.Printf("Replication timeout for key %s", key)
	}

	mu.Lock()
	defer mu.Unlock()

	// Quorum is 1 replica ACK (since primary wrote locally, total 2)
	if ackCount < 1 && len(replicas) > 0 {
		return nil, log.Output(2, "replication quorum failed: 0 acks received")
	}

	return replicatedTo, nil
}
