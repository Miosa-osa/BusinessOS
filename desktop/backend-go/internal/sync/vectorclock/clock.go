package vectorclock

import (
	"encoding/json"
	"fmt"
)

// VectorClock implements a vector clock for distributed event ordering and conflict detection.
// A vector clock is a data structure used for determining the partial ordering of events
// in a distributed system and detecting causality violations.
//
// Each node (BusinessOS instance, OSA instance) maintains a logical clock that increments
// on each local event. When events are synchronized, vector clocks are compared to detect:
// - Causally ordered events (one happened before the other)
// - Concurrent events (neither happened before the other - conflict!)
//
// References:
// - Lamport, L. (1978). "Time, clocks, and the ordering of events in a distributed system"
// - Fidge, C. (1988). "Timestamps in Message-Passing Systems That Preserve the Partial Ordering"
type VectorClock struct {
	clock map[string]int
}

// New creates a new vector clock.
func New() *VectorClock {
	return &VectorClock{
		clock: make(map[string]int),
	}
}

// FromMap creates a vector clock from a map representation.
func FromMap(m map[string]int) *VectorClock {
	clock := make(map[string]int, len(m))
	for k, v := range m {
		clock[k] = v
	}
	return &VectorClock{clock: clock}
}

// FromJSON deserializes a vector clock from JSON.
func FromJSON(data []byte) (*VectorClock, error) {
	var m map[string]int
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("unmarshal vector clock: %w", err)
	}
	return FromMap(m), nil
}

// Increment increments the clock for the specified node ID.
// This should be called when a local event occurs (e.g., creating or updating an entity).
func (vc *VectorClock) Increment(nodeID string) {
	vc.clock[nodeID]++
}

// Get returns the clock value for the specified node ID.
// Returns 0 if the node ID doesn't exist in the clock.
func (vc *VectorClock) Get(nodeID string) int {
	return vc.clock[nodeID]
}

// Set sets the clock value for the specified node ID.
func (vc *VectorClock) Set(nodeID string, value int) {
	vc.clock[nodeID] = value
}

// Merge merges another vector clock into this one, taking the maximum value
// for each node. This is used when receiving an event from another node.
//
// After merging, increment the local node's clock to reflect that a new
// local event (receiving the remote event) has occurred.
func (vc *VectorClock) Merge(other *VectorClock) {
	for nodeID, otherValue := range other.clock {
		if currentValue, exists := vc.clock[nodeID]; !exists || otherValue > currentValue {
			vc.clock[nodeID] = otherValue
		}
	}
}

// Compare compares this vector clock with another and returns:
//   - -1 if this clock is strictly less than the other (this happened before other)
//   - 0 if the clocks are concurrent (conflict - neither happened before the other)
//   - 1 if this clock is strictly greater than the other (this happened after other)
//
// Two clocks are concurrent if neither is strictly less than the other,
// which indicates a conflict that needs resolution.
func (vc *VectorClock) Compare(other *VectorClock) int {
	hasLess := false
	hasGreater := false

	// Get all node IDs from both clocks
	allNodes := make(map[string]bool)
	for nodeID := range vc.clock {
		allNodes[nodeID] = true
	}
	for nodeID := range other.clock {
		allNodes[nodeID] = true
	}

	// Compare each node's timestamp
	for nodeID := range allNodes {
		thisValue := vc.Get(nodeID)
		otherValue := other.Get(nodeID)

		if thisValue < otherValue {
			hasLess = true
		} else if thisValue > otherValue {
			hasGreater = true
		}
	}

	// Determine relationship
	if hasLess && hasGreater {
		return 0 // Concurrent (conflict)
	} else if hasLess {
		return -1 // This happened before other
	} else if hasGreater {
		return 1 // This happened after other
	}

	// Equal (same events, or both are empty)
	// Treat equal clocks as concurrent to be safe
	return 0
}

// IsBefore returns true if this clock happened strictly before the other clock.
func (vc *VectorClock) IsBefore(other *VectorClock) bool {
	return vc.Compare(other) == -1
}

// IsAfter returns true if this clock happened strictly after the other clock.
func (vc *VectorClock) IsAfter(other *VectorClock) bool {
	return vc.Compare(other) == 1
}

// IsConcurrent returns true if this clock is concurrent with the other clock.
// Concurrent clocks indicate a conflict that needs resolution.
func (vc *VectorClock) IsConcurrent(other *VectorClock) bool {
	return vc.Compare(other) == 0
}

// ToMap returns a map representation of the vector clock.
func (vc *VectorClock) ToMap() map[string]int {
	result := make(map[string]int, len(vc.clock))
	for k, v := range vc.clock {
		result[k] = v
	}
	return result
}

// ToJSON serializes the vector clock to JSON.
func (vc *VectorClock) ToJSON() ([]byte, error) {
	return json.Marshal(vc.clock)
}

// Clone creates a deep copy of the vector clock.
func (vc *VectorClock) Clone() *VectorClock {
	return FromMap(vc.ToMap())
}

// String returns a string representation of the vector clock for debugging.
func (vc *VectorClock) String() string {
	data, _ := vc.ToJSON()
	return string(data)
}
