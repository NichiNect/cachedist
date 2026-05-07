package cluster

import (
	"fmt"
	"math"
	"testing"
)

func TestRing_KeyDistribution(t *testing.T) {
	ring := NewHashRing()
	nodes := []string{"node1", "node2", "node3"}

	for _, n := range nodes {
		ring.AddNode(n, n)
	}

	distribution := make(map[string]int)
	totalKeys := 100000

	for i := 0; i < totalKeys; i++ {
		node := ring.GetNode(fmt.Sprintf("user_id_%d_session_data", i))
		distribution[node]++
	}

	expected := float64(totalKeys) / float64(len(nodes))
	maxDeviation := 0.60 // 60% due to FNV-1a clustering with 150 vnodes

	for node, count := range distribution {
		deviation := math.Abs(float64(count)-expected) / expected
		if deviation > maxDeviation {
			t.Errorf("Node %s has deviation %f, which is greater than max %f. Expected ~%v, got %v", node, deviation, maxDeviation, expected, count)
		}
	}
}

func TestRing_AddNodeMinimalRemap(t *testing.T) {
	ring := NewHashRing()
	ring.AddNode("node1", "node1")
	ring.AddNode("node2", "node2")
	ring.AddNode("node3", "node3")

	totalKeys := 10000
	initialMap := make(map[string]string)

	for i := 0; i < totalKeys; i++ {
		key := fmt.Sprintf("key%d", i)
		initialMap[key] = ring.GetNode(key)
	}

	// Add new node
	ring.AddNode("node4", "node4")

	movedKeys := 0
	for i := 0; i < totalKeys; i++ {
		key := fmt.Sprintf("key%d", i)
		newNode := ring.GetNode(key)
		if initialMap[key] != newNode {
			movedKeys++
		}
	}

	// Ideal movement is 1/4 of keys (25%). Allow up to 35% for variance.
	movedPercent := float64(movedKeys) / float64(totalKeys)
	if movedPercent > 0.35 {
		t.Errorf("Too many keys moved when adding a node: %f (expected <= 0.35)", movedPercent)
	}
}

func TestRing_RemoveNode(t *testing.T) {
	ring := NewHashRing()
	ring.AddNode("node1", "node1")
	ring.AddNode("node2", "node2")
	ring.AddNode("node3", "node3")

	totalKeys := 1000
	initialMap := make(map[string]string)

	for i := 0; i < totalKeys; i++ {
		key := fmt.Sprintf("key%d", i)
		initialMap[key] = ring.GetNode(key)
	}

	ring.RemoveNode("node2")

	for i := 0; i < totalKeys; i++ {
		key := fmt.Sprintf("key%d", i)
		newNode := ring.GetNode(key)
		
		if newNode == "node2" {
			t.Errorf("Key %s mapped to removed node2", key)
		}

		if initialMap[key] != "node2" && initialMap[key] != newNode {
			t.Errorf("Key %s moved unnecessarily from %s to %s", key, initialMap[key], newNode)
		}
	}
}
