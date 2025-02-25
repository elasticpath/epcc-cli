package toposort

import "fmt"

func NewGraph() *Graph {
	return &Graph{
		edges:    make(map[string][]string),
		vertices: make(map[string]struct{}),
	}
}

// Adapted from: https://reintech.io/blog/topological-sorting-in-go
type Graph struct {
	edges    map[string][]string
	vertices map[string]struct{}
}

func (g *Graph) AddEdge(u, v string) {
	g.vertices[u] = struct{}{}
	g.vertices[v] = struct{}{}

	g.edges[u] = append(g.edges[u], v)
}

func (g *Graph) topologicalSortUtil(v string, visited map[string]bool, inStack map[string]bool, stack *[]string) error {
	visited[v] = true
	inStack[v] = true // Mark node as being in the recursion stack

	for _, u := range g.edges[v] {
		if !visited[u] {
			if err := g.topologicalSortUtil(u, visited, inStack, stack); err != nil {
				return err
			}
		} else if inStack[u] {
			return fmt.Errorf("Cycle Detected Between Edges, %s and %s", v, u)
		}
	}

	inStack[v] = false // Remove from recursion stack after processing
	*stack = append([]string{v}, *stack...)
	return nil
}

func (g *Graph) TopologicalSort() ([]string, error) {
	stack := []string{}
	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	for v := range g.vertices {
		if !visited[v] {
			if err := g.topologicalSortUtil(v, visited, inStack, &stack); err != nil {
				return nil, err
			}
		}
	}

	return stack, nil
}
