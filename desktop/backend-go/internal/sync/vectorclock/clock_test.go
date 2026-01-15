package vectorclock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	vc := New()
	assert.NotNil(t, vc)
	assert.Empty(t, vc.clock)
}

func TestIncrement(t *testing.T) {
	vc := New()
	vc.Increment("node1")
	assert.Equal(t, 1, vc.Get("node1"))

	vc.Increment("node1")
	assert.Equal(t, 2, vc.Get("node1"))

	vc.Increment("node2")
	assert.Equal(t, 1, vc.Get("node2"))
}

func TestGet(t *testing.T) {
	vc := New()
	vc.Set("node1", 5)

	assert.Equal(t, 5, vc.Get("node1"))
	assert.Equal(t, 0, vc.Get("nonexistent"))
}

func TestMerge(t *testing.T) {
	vc1 := New()
	vc1.Set("node1", 3)
	vc1.Set("node2", 1)

	vc2 := New()
	vc2.Set("node1", 2)
	vc2.Set("node2", 4)
	vc2.Set("node3", 1)

	vc1.Merge(vc2)

	// Should take maximum values
	assert.Equal(t, 3, vc1.Get("node1")) // kept vc1's higher value
	assert.Equal(t, 4, vc1.Get("node2")) // took vc2's higher value
	assert.Equal(t, 1, vc1.Get("node3")) // added new node from vc2
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		clock1   map[string]int
		clock2   map[string]int
		expected int
	}{
		{
			name:     "clock1 before clock2",
			clock1:   map[string]int{"node1": 1, "node2": 2},
			clock2:   map[string]int{"node1": 2, "node2": 3},
			expected: -1,
		},
		{
			name:     "clock1 after clock2",
			clock1:   map[string]int{"node1": 3, "node2": 4},
			clock2:   map[string]int{"node1": 2, "node2": 3},
			expected: 1,
		},
		{
			name:     "concurrent clocks",
			clock1:   map[string]int{"node1": 3, "node2": 1},
			clock2:   map[string]int{"node1": 2, "node2": 4},
			expected: 0,
		},
		{
			name:     "equal clocks",
			clock1:   map[string]int{"node1": 2, "node2": 3},
			clock2:   map[string]int{"node1": 2, "node2": 3},
			expected: 0,
		},
		{
			name:     "empty clocks",
			clock1:   map[string]int{},
			clock2:   map[string]int{},
			expected: 0,
		},
		{
			name:     "one empty clock",
			clock1:   map[string]int{"node1": 1},
			clock2:   map[string]int{},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vc1 := FromMap(tt.clock1)
			vc2 := FromMap(tt.clock2)
			result := vc1.Compare(vc2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsBefore(t *testing.T) {
	vc1 := FromMap(map[string]int{"node1": 1, "node2": 2})
	vc2 := FromMap(map[string]int{"node1": 2, "node2": 3})

	assert.True(t, vc1.IsBefore(vc2))
	assert.False(t, vc2.IsBefore(vc1))
}

func TestIsAfter(t *testing.T) {
	vc1 := FromMap(map[string]int{"node1": 3, "node2": 4})
	vc2 := FromMap(map[string]int{"node1": 2, "node2": 3})

	assert.True(t, vc1.IsAfter(vc2))
	assert.False(t, vc2.IsAfter(vc1))
}

func TestIsConcurrent(t *testing.T) {
	vc1 := FromMap(map[string]int{"node1": 3, "node2": 1})
	vc2 := FromMap(map[string]int{"node1": 2, "node2": 4})

	assert.True(t, vc1.IsConcurrent(vc2))
	assert.True(t, vc2.IsConcurrent(vc1))
}

func TestFromMap(t *testing.T) {
	m := map[string]int{"node1": 5, "node2": 3}
	vc := FromMap(m)

	assert.Equal(t, 5, vc.Get("node1"))
	assert.Equal(t, 3, vc.Get("node2"))

	// Verify it's a copy, not a reference
	m["node1"] = 10
	assert.Equal(t, 5, vc.Get("node1"))
}

func TestToMap(t *testing.T) {
	vc := New()
	vc.Set("node1", 5)
	vc.Set("node2", 3)

	m := vc.ToMap()
	assert.Equal(t, 5, m["node1"])
	assert.Equal(t, 3, m["node2"])

	// Verify it's a copy, not a reference
	m["node1"] = 10
	assert.Equal(t, 5, vc.Get("node1"))
}

func TestJSON(t *testing.T) {
	vc := New()
	vc.Set("node1", 5)
	vc.Set("node2", 3)

	// Serialize
	data, err := vc.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Deserialize
	vc2, err := FromJSON(data)
	assert.NoError(t, err)
	assert.Equal(t, 5, vc2.Get("node1"))
	assert.Equal(t, 3, vc2.Get("node2"))
}

func TestClone(t *testing.T) {
	vc1 := New()
	vc1.Set("node1", 5)
	vc1.Set("node2", 3)

	vc2 := vc1.Clone()
	assert.Equal(t, 5, vc2.Get("node1"))
	assert.Equal(t, 3, vc2.Get("node2"))

	// Verify it's a deep copy
	vc1.Set("node1", 10)
	assert.Equal(t, 10, vc1.Get("node1"))
	assert.Equal(t, 5, vc2.Get("node1"))
}

func TestString(t *testing.T) {
	vc := New()
	vc.Set("node1", 5)
	vc.Set("node2", 3)

	str := vc.String()
	assert.NotEmpty(t, str)
	assert.Contains(t, str, "node1")
	assert.Contains(t, str, "node2")
}
