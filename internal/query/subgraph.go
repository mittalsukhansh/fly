package query

import (
	"github.com/fly/connectomegraph/internal/graph"
)

// SubgraphResult represents the extracted subgraph nodes and edges.
type SubgraphResult struct {
	Nodes []graph.NeuronID
	Edges []graph.Edge // Note: Edge.To is the destination, but we need the Source as well. We should use a custom struct.
}

type SubgraphEdge struct {
	From   graph.NeuronID
	To     graph.NeuronID
	Weight uint16
}

type SubgraphResponse struct {
	Nodes []graph.NeuronID
	Edges []SubgraphEdge
}

// ExtractSubgraphByType filters nodes by a specific metadata Type and
// returns the induced subgraph (only edges where BOTH endpoints match the type).
func ExtractSubgraphByType(g *graph.Graph, neuronType string) SubgraphResponse {
	var nodes []graph.NeuronID
	inSubgraph := make(map[graph.NeuronID]bool)

	// 1. Find all matching nodes
	for id, meta := range g.Meta {
		if meta.Type == neuronType {
			nID := graph.NeuronID(id)
			nodes = append(nodes, nID)
			inSubgraph[nID] = true
		}
	}

	var edges []SubgraphEdge
	
	// 2. Extract edges between matching nodes
	for _, nID := range nodes {
		for _, edge := range g.OutEdges(nID) {
			if inSubgraph[edge.To] {
				edges = append(edges, SubgraphEdge{
					From:   nID,
					To:     edge.To,
					Weight: edge.Weight,
				})
			}
		}
	}

	return SubgraphResponse{
		Nodes: nodes,
		Edges: edges,
	}
}
