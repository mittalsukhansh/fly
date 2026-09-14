package query

import (
	"github.com/fly/connectomegraph/internal/graph"
)

// Neighborhood extracts an N-hop neighborhood for a given node.
// direction can be "up" (incoming edges) or "down" (outgoing edges).
// Returns a map of visited NeuronIDs.
func Neighborhood(g *graph.Graph, start graph.NeuronID, depth int, direction string) map[graph.NeuronID]bool {
	visited := make(map[graph.NeuronID]bool)
	visited[start] = true

	queue := []graph.NeuronID{start}
	
	// Track depth per node to stop at N hops
	nodeDepths := make(map[graph.NeuronID]int)
	nodeDepths[start] = 0

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		d := nodeDepths[curr]
		if d >= depth {
			continue // stop traversing from this node
		}

		var edges []graph.Edge
		if direction == "up" {
			edges = g.InEdges(curr)
		} else {
			edges = g.OutEdges(curr)
		}

		for _, edge := range edges {
			if !visited[edge.To] {
				visited[edge.To] = true
				nodeDepths[edge.To] = d + 1
				queue = append(queue, edge.To)
			}
		}
	}

	return visited
}
