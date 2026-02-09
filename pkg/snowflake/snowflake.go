package snowflake

import (
	"errors"
	"sync"
	"time"
)

// Config holds the configuration for our node
const (
	epoch     = int64(1672531200000) // Custom Epoch (Jan 1, 2023) to keep IDs smaller
	nodeBits  = 10                   // Number of bits for Node ID (0-1023 nodes)
	stepBits  = 12                   // Number of bits for Step (4096 IDs per ms)
	nodeMax   = -1 ^ (-1 << nodeBits)
	stepMax   = -1 ^ (-1 << stepBits)
	timeShift = nodeBits + stepBits
	nodeShift = stepBits
)

// Node is our ID generator worker
type Node struct {
	mu        sync.Mutex // A lock to ensure safety in concurrent requests
	timestamp int64
	nodeID    int64
	step      int64
}

// NewNode creates a new generator for a specific machine (0-1023)
func NewNode(nodeID int64) (*Node, error) {
	if nodeID < 0 || nodeID > nodeMax {
		return nil, errors.New("node ID too large")
	}

	return &Node{
		timestamp: 0,
		nodeID:    nodeID,
		step:      0,
	}, nil
}

// Generate creates a unique ID
func (n *Node) Generate() int64 {
	n.mu.Lock()         // Lock the door! Only one ID generated at a time.
	defer n.mu.Unlock() // Unlock when function finishes.

	now := time.Now().UnixMilli() // Current time in ms

	if now == n.timestamp {
		// If we are in the same millisecond, increment the step
		n.step = (n.step + 1) & stepMax
		if n.step == 0 {
			// We ran out of steps for this millisecond! Wait for next one.
			for now <= n.timestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		// New millisecond, reset step
		n.step = 0
	}

	n.timestamp = now

	// The Magic: Shift bits to combine Time + Node + Step into one number
	id := ((now - epoch) << timeShift) | (n.nodeID << nodeShift) | n.step
	return id
}
