package toposort

import (
	"fmt"
	"sort"
)

type Graph struct {
	edges    map[string][]string
	vertices map[string]struct{}
	indegree map[string]int // Track number of dependencies
}

func NewGraph() *Graph {
	return &Graph{
		edges:    make(map[string][]string),
		vertices: make(map[string]struct{}),
		indegree: make(map[string]int),
	}
}

func (g *Graph) AddNode(v string) {
	g.vertices[v] = struct{}{}
	if _, exists := g.edges[v]; !exists {
		g.edges[v] = []string{} // Ensure it exists in the adjacency list
	}
}

func (g *Graph) AddEdge(u, v string) {
	g.vertices[u] = struct{}{}
	g.vertices[v] = struct{}{}
	g.edges[u] = append(g.edges[u], v)
	g.indegree[v]++ // Track dependencies
}

// Determines parallel execution stages using Kahn's Algorithm
func (g *Graph) ParallelizableStages() ([][]string, error) {
	indegree := make(map[string]int)
	for v := range g.vertices {
		indegree[v] = g.indegree[v] // Copy original indegrees
	}

	queue := []string{}
	for v, deg := range indegree {
		if deg == 0 {
			queue = append(queue, v)
		}
	}

	var levels [][]string
	count := 0

	for len(queue) > 0 {
		sort.Strings(queue) // Sort the current stage before processing
		var nextQueue []string
		levels = append(levels, queue) // Nodes at current level
		count += len(queue)

		for _, v := range queue {
			for _, neighbor := range g.edges[v] {
				indegree[neighbor]--
				if indegree[neighbor] == 0 {
					nextQueue = append(nextQueue, neighbor)
				}
			}
		}

		queue = nextQueue // Move to next level
	}

	// If not all nodes were processed, there is a cycle
	if count != len(g.vertices) {
		var cycleNodes []string
		for v, deg := range indegree {
			if deg > 0 {
				cycleNodes = append(cycleNodes, v)
			}
		}
		sort.Strings(cycleNodes)

		return nil, fmt.Errorf("cycle detected in graph : %v", cycleNodes)
	}

	return levels, nil
}

// Topological Sort - Simply flattens ParallelizableStages output
func (g *Graph) TopologicalSort() ([]string, error) {
	stages, err := g.ParallelizableStages()
	if err != nil {
		return nil, err
	}

	var order []string
	for _, stage := range stages {
		order = append(order, stage...)
	}

	return order, nil
}
