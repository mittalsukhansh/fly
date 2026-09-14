package graph

// GraphBuilder helps construct a CSR Graph efficiently.
// We use a two-pass approach to avoid holding all raw edges in memory.
// Pass 1: Count degrees to pre-allocate arrays.
// Pass 2: Place edges directly into their final CSR positions.
type GraphBuilder struct {
	maxNodeID  uint32
	outDegrees []uint32
	inDegrees  []uint32
	totalEdges uint32

	// Final graph being built
	graph *Graph

	// Temporary offsets used during Pass 2
	currentOutOffsets []uint32
	currentInOffsets  []uint32
}

// NewGraphBuilder initializes the builder. maxNodeID can be an estimate,
// the slices will grow if necessary.
func NewGraphBuilder(estimatedMaxNodeID uint32) *GraphBuilder {
	return &GraphBuilder{
		maxNodeID:  0,
		outDegrees: make([]uint32, estimatedMaxNodeID+1),
		inDegrees:  make([]uint32, estimatedMaxNodeID+1),
	}
}

func (b *GraphBuilder) ensureCapacity(nodeID uint32) {
	if nodeID > b.maxNodeID {
		b.maxNodeID = nodeID
	}
	for uint32(len(b.outDegrees)) <= nodeID {
		b.outDegrees = append(b.outDegrees, 0)
		b.inDegrees = append(b.inDegrees, 0)
	}
}

// AddEdgePass1 increments the degree counts for a given edge.
func (b *GraphBuilder) AddEdgePass1(from, to NeuronID, weight uint16) {
	b.ensureCapacity(uint32(from))
	b.ensureCapacity(uint32(to))
	b.outDegrees[from]++
	b.inDegrees[to]++
	b.totalEdges++
}

// Allocate initializes the CSR arrays after Pass 1 is complete.
func (b *GraphBuilder) Allocate(meta []NeuronMeta) {
	numNodes := b.maxNodeID + 1

	g := &Graph{
		Offsets:    make([]uint32, numNodes+1),
		Edges:      make([]Edge, b.totalEdges),
		RevOffsets: make([]uint32, numNodes+1),
		RevEdges:   make([]Edge, b.totalEdges),
		Meta:       meta,
	}

	b.currentOutOffsets = make([]uint32, numNodes+1)
	b.currentInOffsets = make([]uint32, numNodes+1)

	var outSum, inSum uint32
	for i := uint32(0); i < numNodes; i++ {
		g.Offsets[i] = outSum
		b.currentOutOffsets[i] = outSum
		outSum += b.outDegrees[i]

		g.RevOffsets[i] = inSum
		b.currentInOffsets[i] = inSum
		inSum += b.inDegrees[i]
	}
	g.Offsets[numNodes] = outSum
	g.RevOffsets[numNodes] = inSum

	b.graph = g
}

// AddEdgePass2 places an edge into the pre-allocated CSR arrays.
func (b *GraphBuilder) AddEdgePass2(from, to NeuronID, weight uint16) {
	// Forward
	outPos := b.currentOutOffsets[from]
	b.graph.Edges[outPos] = Edge{To: to, Weight: weight}
	b.currentOutOffsets[from]++

	// Reverse
	inPos := b.currentInOffsets[to]
	b.graph.RevEdges[inPos] = Edge{To: from, Weight: weight} // Note: To stores the 'from' node in RevEdges
	b.currentInOffsets[to]++
}

// Build returns the finalized Graph.
func (b *GraphBuilder) Build() *Graph {
	return b.graph
}
