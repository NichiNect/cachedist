package bench

import (
	"fmt"
	"math/rand"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/NichiNect/cachedist/sdk"
)

func TestChaosScenario(t *testing.T) {
	// 1. Setup Client
	addrs := []string{"localhost:7001", "localhost:7002", "localhost:7003"}
	client := sdk.NewClient(addrs)

	stopChan := make(chan struct{})
	var wg sync.WaitGroup
	
	errors := 0
	total := 0
	var mu sync.Mutex

	// 2. Start constant background load
	fmt.Println("🚀 Starting background load...")
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
			for {
				select {
				case <-stopChan:
					return
				default:
					key := fmt.Sprintf("chaos_key_%d", r.Intn(1000))
					err := client.Set(key, "value", 10)
					
					mu.Lock()
					total++
					if err != nil {
						errors++
					}
					mu.Unlock()
					
					time.Sleep(10 * time.Millisecond)
					client.Get(key)
				}
			}
		}(i)
	}

	// 3. Chaos Phase: Kill Node 2
	time.Sleep(3 * time.Second)
	fmt.Println("🔥 CHAOS: Killing node2...")
	exec.Command("docker", "stop", "node2").Run()

	// Wait while node is down
	time.Sleep(5 * time.Second)
	
	mu.Lock()
	fmt.Printf("📊 Stats while node2 is DOWN: Total: %d, Errors: %d (Rate: %.2f%%)\n", 
		total, errors, float64(errors)/float64(total)*100)
	mu.Unlock()

	// 4. Recovery Phase: Start Node 2 back up
	fmt.Println("🛠 RECOVERY: Bringing node2 back online...")
	exec.Command("docker", "start", "node2").Run()
	
	// Wait for node to rejoin and sync
	time.Sleep(10 * time.Second)

	// 5. Final validation
	close(stopChan)
	wg.Wait()

	fmt.Printf("✅ Final Results: Total Ops: %d, Total Errors: %d\n", total, errors)
	if float64(errors)/float64(total) > 0.1 {
		t.Errorf("Too many errors during chaos! Error rate: %.2f%%", float64(errors)/float64(total)*100)
	} else {
		fmt.Println("🌟 System survived chaos with acceptable error rate!")
	}
}
