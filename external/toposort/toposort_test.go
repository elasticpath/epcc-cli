package toposort

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTopoSort(t *testing.T) {
	// Fixture Setup
	g := NewGraph()
	// Execute SUT
	g.AddEdge("c", "b")
	g.AddEdge("b", "a")

	sort, err := g.TopologicalSort()

	// Verification
	require.NoError(t, err)
	require.Equal(t, []string{"c", "b", "a"}, sort)
}

func TestTopoSortCircle(t *testing.T) {
	// Fixture Setup
	g := NewGraph()
	// Execute SUT
	g.AddEdge("c", "b")
	g.AddEdge("b", "a")
	g.AddEdge("a", "c")

	_, err := g.TopologicalSort()

	// Verification
	require.ErrorContains(t, err, "Cycle Detected Between Edges")
}
