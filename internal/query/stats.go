package query

import (
	"sort"

	"github.com/fly/connectomegraph/internal/graph"
)

type DegreeStats struct {
	TotalNodes uint32
	TotalEdges int
}

type NodeDegree struct {
	ID        graph.NeuronID
	InDegree  int
	OutDegree int
	Total     int
}

// GetBasicStats returns graph-level statistics.
func GetBasicStats(g *graph.Graph) DegreeStats {
	return DegreeStats{
		TotalNodes: g.MaxNodeID() + 1,
		TotalEdges: len(g.Edges),
	}
}

// GetTopConnected returns the top N nodes by total degree (in + out).
func GetTopConnected(g *graph.Graph, n int) []NodeDegree {
	maxID := g.MaxNodeID()
	var degrees []NodeDegree

	for id := uint32(0); id <= maxID; id++ {
		nID := graph.NeuronID(id)
		outD := len(g.OutEdges(nID))
		inD := len(g.InEdges(nID))
		if outD > 0 || inD > 0 {
			degrees = append(degrees, NodeDegree{
				ID:        nID,
				InDegree:  inD,
				OutDegree: outD,
				Total:     inD + outD,
			})
		}
	}

	sort.Slice(degrees, func(i, j int) bool {
		return degrees[i].Total > degrees[j].Total
	})

	if len(degrees) > n {
		return degrees[:n]
	}
	return degrees
}
