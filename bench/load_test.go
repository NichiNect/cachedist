package bench

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/NichiNect/cachedist/sdk"
)

func TestLoadScenario(t *testing.T) {
	// Addresses for the 3-node cluster
	addrs := []string{"localhost:7001", "localhost:7002", "localhost:7003"}
	client := sdk.NewClient(addrs)

	const (
		numGoroutines = 10
		totalOps      = 100000
		opsPerG       = totalOps / numGoroutines
	)

	var wg sync.WaitGroup
	start := time.Now()

	errCount := 0
	var mu sync.Mutex

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
			
			localErrCount := 0
			for i := 0; i < opsPerG; i++ {
				key := fmt.Sprintf("load_key_%d", r.Intn(100000))
				op := r.Float64()

				var err error
				if op < 0.7 {
					_, _, err = client.Get(key)
				} else if op < 0.9 {
					err = client.Set(key, "load_value", 0)
				} else {
					err = client.Delete(key)
				}

				if err != nil {
					localErrCount++
				}
			}

			mu.Lock()
			errCount += localErrCount
			mu.Unlock()
		}(g)
	}

	wg.Wait()
	duration := time.Since(start)

	throughput := float64(totalOps) / duration.Seconds()
	fmt.Printf("Load Test Results:\n")
	fmt.Printf("Total Ops: %d\n", totalOps)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Throughput: %.2f ops/sec\n", throughput)
	fmt.Printf("Error Rate: %.2f%%\n", float64(errCount)/float64(totalOps)*100)
}
