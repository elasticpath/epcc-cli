package toposort

import (
	"testing"

	"github.com/stretchr/testify/require"
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

func TestTopoSortWithExtraNodes(t *testing.T) {
	// Fixture Setup
	g := NewGraph()
	// Execute SUT
	g.AddNode("c")
	g.AddEdge("c", "b")
	g.AddNode("d")
	g.AddEdge("b", "a")
	g.AddNode("a")

	sort, err := g.TopologicalSort()

	// Verification
	require.NoError(t, err)
	require.Equal(t, []string{"c", "d", "b", "a"}, sort)
}

func TestTopoSortWithStagesAndExtraNodes(t *testing.T) {
	// Fixture Setup
	g := NewGraph()
	// Execute SUT
	g.AddNode("c")
	g.AddEdge("c", "b")
	g.AddNode("d")
	g.AddEdge("b", "a")
	g.AddNode("a")
	g.AddEdge("b", "x")
	g.AddEdge("y", "z")
	g.AddEdge("x", "z")
	g.AddEdge("f", "y")
	g.AddEdge("w", "x")

	sort, err := g.ParallelizableStages()

	// Verification
	require.NoError(t, err)
	require.Equal(t, [][]string{{"c", "d", "f", "w"}, {"b", "y"}, {"a", "x"}, {"z"}}, sort)
}

func TestTopoSortCircle(t *testing.T) {
	// Fixture Setup
	g := NewGraph()

	// Execute SUT
	g.AddEdge("e", "c")
	g.AddEdge("c", "b")
	g.AddEdge("b", "a")
	g.AddEdge("a", "c")
	g.AddEdge("d", "b")
	g.AddEdge("c", "f")

	_, err := g.TopologicalSort()

	// Verification
	require.ErrorContains(t, err, "cycle detected in graph")
}
