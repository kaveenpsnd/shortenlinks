package snowflake

import (
	"sync"
	"testing"
)

func TestNewNode_ValidID(t *testing.T) {
	node, err := NewNode(1)
	if err != nil {
		t.Fatalf("NewNode(1) returned error: %v", err)
	}
	if node == nil {
		t.Fatal("NewNode(1) returned nil node")
	}
}

func TestNewNode_ZeroID(t *testing.T) {
	node, err := NewNode(0)
	if err != nil {
		t.Fatalf("NewNode(0) returned error: %v", err)
	}
	if node == nil {
		t.Fatal("NewNode(0) returned nil node")
	}
}

func TestNewNode_MaxID(t *testing.T) {
	node, err := NewNode(int64(nodeMax))
	if err != nil {
		t.Fatalf("NewNode(%d) returned error: %v", nodeMax, err)
	}
	if node == nil {
		t.Fatal("NewNode(max) returned nil node")
	}
}

func TestNewNode_NegativeID(t *testing.T) {
	_, err := NewNode(-1)
	if err == nil {
		t.Error("NewNode(-1) should return an error")
	}
}

func TestNewNode_OverMaxID(t *testing.T) {
	_, err := NewNode(int64(nodeMax) + 1)
	if err == nil {
		t.Errorf("NewNode(%d) should return an error", nodeMax+1)
	}
}

func TestGenerate_ReturnsPositiveID(t *testing.T) {
	node, _ := NewNode(1)
	id := node.Generate()
	if id <= 0 {
		t.Errorf("Generate() returned non-positive ID: %d", id)
	}
}

func TestGenerate_UniqueIDs(t *testing.T) {
	node, _ := NewNode(1)
	ids := make(map[int64]bool)

	for i := 0; i < 10000; i++ {
		id := node.Generate()
		if ids[id] {
			t.Fatalf("Duplicate ID generated: %d at iteration %d", id, i)
		}
		ids[id] = true
	}
}

func TestGenerate_Increasing(t *testing.T) {
	node, _ := NewNode(1)
	var prev int64

	for i := 0; i < 1000; i++ {
		id := node.Generate()
		if id <= prev {
			t.Errorf("ID %d is not greater than previous %d at iteration %d", id, prev, i)
		}
		prev = id
	}
}

func TestGenerate_ConcurrentSafety(t *testing.T) {
	node, _ := NewNode(1)
	const numGoroutines = 100
	const idsPerGoroutine = 100

	var mu sync.Mutex
	ids := make(map[int64]bool)
	var wg sync.WaitGroup

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localIDs := make([]int64, idsPerGoroutine)
			for i := 0; i < idsPerGoroutine; i++ {
				localIDs[i] = node.Generate()
			}

			mu.Lock()
			for _, id := range localIDs {
				if ids[id] {
					t.Errorf("Duplicate ID in concurrent generation: %d", id)
				}
				ids[id] = true
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	expected := numGoroutines * idsPerGoroutine
	if len(ids) != expected {
		t.Errorf("Expected %d unique IDs, got %d", expected, len(ids))
	}
}

func TestGenerate_DifferentNodes(t *testing.T) {
	node1, _ := NewNode(1)
	node2, _ := NewNode(2)

	id1 := node1.Generate()
	id2 := node2.Generate()

	if id1 == id2 {
		t.Errorf("Different nodes generated the same ID: %d", id1)
	}
}

func BenchmarkGenerate(b *testing.B) {
	node, _ := NewNode(1)
	for i := 0; i < b.N; i++ {
		node.Generate()
	}
}
