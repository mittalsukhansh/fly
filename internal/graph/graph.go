package graph

import (
	"fmt"
)

// NeuronID represents a unique identifier for a neuron. 166,700 fits comfortably in uint32.
type NeuronID uint32

// Edge represents a directed connection to another neuron with a specific weight (synapse count).
type Edge struct {
	To     NeuronID
	Weight uint16
}

// NeuronMeta holds metadata about a neuron.
type NeuronMeta struct {
	Type   string
	X      float32
	Y      float32
	Z      float32
	EmbedX float32
	EmbedY float32
	EmbedZ float32
}

// Graph is a compressed sparse row (CSR) representation of the connectome.
type Graph struct {
	// Forward CSR: outgoing edges
	Offsets []uint32 // len = numNodes + 1
	Edges   []Edge   // len = numEdges, sorted by source node

	// Reverse CSR: incoming edges (built alongside forward for upstream queries)
	RevOffsets []uint32
	RevEdges   []Edge

	// Node metadata indexed by NeuronID
	Meta []NeuronMeta // len = numNodes
}

// MaxNodeID returns the maximum node ID based on the offset array length.
func (g *Graph) MaxNodeID() uint32 {
	if len(g.Offsets) == 0 {
		return 0
	}
	return uint32(len(g.Offsets) - 1)
}

// OutEdges returns the outgoing edges for a given neuron.
func (g *Graph) OutEdges(node NeuronID) []Edge {
	if uint32(node) >= g.MaxNodeID() {
		return nil
	}
	start := g.Offsets[node]
	end := g.Offsets[node+1]
	return g.Edges[start:end]
}

// InEdges returns the incoming edges for a given neuron using the reverse CSR.
func (g *Graph) InEdges(node NeuronID) []Edge {
	if uint32(node) >= g.MaxNodeID() {
		return nil
	}
	start := g.RevOffsets[node]
	end := g.RevOffsets[node+1]
	return g.RevEdges[start:end]
}

// PrintStats prints basic graph statistics.
func (g *Graph) PrintStats() {
	fmt.Printf("Graph Stats:\n")
	fmt.Printf("  Nodes: %d\n", g.MaxNodeID())
	fmt.Printf("  Edges: %d\n", len(g.Edges))
}
