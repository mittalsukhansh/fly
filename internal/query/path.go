package query

import (
	"container/heap"
	"math"

	"github.com/fly/connectomegraph/internal/graph"
)

// ShortestPathHopCount uses BFS to find the shortest path between `from` and `to`.
// Returns the path of NeuronIDs and true if found, nil and false otherwise.
func ShortestPathHopCount(g *graph.Graph, from, to graph.NeuronID) ([]graph.NeuronID, bool) {
	if from == to {
		return []graph.NeuronID{from}, true
	}

	queue := []graph.NeuronID{from}
	visited := make(map[graph.NeuronID]bool)
	visited[from] = true

	// Track parent to reconstruct the path
	parent := make(map[graph.NeuronID]graph.NeuronID)

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr == to {
			// Reconstruct path
			path := []graph.NeuronID{}
			for n := to; n != from; n = parent[n] {
				path = append(path, n)
			}
			path = append(path, from)
			
			// Reverse
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			return path, true
		}

		for _, edge := range g.OutEdges(curr) {
			if !visited[edge.To] {
				visited[edge.To] = true
				parent[edge.To] = curr
				queue = append(queue, edge.To)
			}
		}
	}

	return nil, false
}

// Item for Dijkstra Priority Queue
type dijkstraItem struct {
	node graph.NeuronID
	cost float64
	idx  int
}

type priorityQueue []*dijkstraItem

func (pq priorityQueue) Len() int { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].cost < pq[j].cost }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].idx = i
	pq[j].idx = j
}
func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*dijkstraItem)
	item.idx = n
	*pq = append(*pq, item)
}
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.idx = -1
	*pq = old[0 : n-1]
	return item
}

// ShortestPathWeighted uses Dijkstra to find the strongest path.
// Cost = 1.0 / weight
func ShortestPathWeighted(g *graph.Graph, from, to graph.NeuronID) ([]graph.NeuronID, bool) {
	if from == to {
		return []graph.NeuronID{from}, true
	}

	dist := make(map[graph.NeuronID]float64)
	parent := make(map[graph.NeuronID]graph.NeuronID)
	
	pq := make(priorityQueue, 0)
	heap.Init(&pq)

	dist[from] = 0
	heap.Push(&pq, &dijkstraItem{node: from, cost: 0})

	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*dijkstraItem)
		u := item.node

		if u == to {
			path := []graph.NeuronID{}
			for n := to; n != from; n = parent[n] {
				path = append(path, n)
			}
			path = append(path, from)
			
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			return path, true
		}

		if item.cost > dist[u] {
			continue
		}

		for _, edge := range g.OutEdges(u) {
			v := edge.To
			// using 1.0 / weight for "strongest path" meaning highest weight is cheapest cost
			weightCost := 1.0 / float64(edge.Weight)
			if edge.Weight == 0 { // Just in case
				weightCost = math.MaxFloat64
			}
			
			alt := dist[u] + weightCost
			
			d, exists := dist[v]
			if !exists || alt < d {
				dist[v] = alt
				parent[v] = u
				heap.Push(&pq, &dijkstraItem{node: v, cost: alt})
			}
		}
	}

	return nil, false
}
