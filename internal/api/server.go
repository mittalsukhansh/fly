package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fly/connectomegraph/internal/graph"
	"github.com/fly/connectomegraph/internal/query"
)

const MaxRenderNodes = 3000 // Result size cap to prevent frontend crash

type EnrichedNode struct {
	ID     uint32  `json:"id"`
	Type   string  `json:"type"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Z      float32 `json:"z"`
	EmbedX float32 `json:"embedX"`
	EmbedY float32 `json:"embedY"`
	EmbedZ float32 `json:"embedZ"`
}

type MinimalNode struct {
	ID     uint32  `json:"id"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Z      float32 `json:"z"`
	EmbedX float32 `json:"embedX"`
	EmbedY float32 `json:"embedY"`
	EmbedZ float32 `json:"embedZ"`
}

type EnrichedEdge struct {
	Source uint32 `json:"source"`
	Target uint32 `json:"target"`
	Weight uint16 `json:"weight"`
}

type GraphResponse struct {
	Nodes []EnrichedNode `json:"nodes"`
	Edges []EnrichedEdge `json:"edges"`
}

type Server struct {
	g *graph.Graph
}

func NewServer(g *graph.Graph) *Server {
	return &Server{g: g}
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	
	// API Endpoints
	mux.HandleFunc("/neuron", s.handleNeuron) // /neuron?id=X
	mux.HandleFunc("/nodes/all", s.handleAllNodes) // /nodes/all
	mux.HandleFunc("/path", s.handlePath)
	mux.HandleFunc("/neighborhood", s.handleNeighborhood) // /neighborhood?id=X&depth=Y&direction=up
	mux.HandleFunc("/subgraph", s.handleSubgraph)
	mux.HandleFunc("/stats/degree", s.handleStatsDegree)
	mux.HandleFunc("/stats/top-connected", s.handleStatsTopConnected)

	// Serve static files for frontend
	fs := http.FileServer(http.Dir("web"))
	mux.Handle("/", fs)

	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleNeuron(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || uint32(id) > s.g.MaxNodeID() {
		http.Error(w, "invalid neuron ID", http.StatusBadRequest)
		return
	}

	meta := s.g.Meta[id]
	json.NewEncoder(w).Encode(meta)
}

func (s *Server) handleAllNodes(w http.ResponseWriter, r *http.Request) {
	nodes := make([]MinimalNode, 0)
	maxID := s.g.MaxNodeID()
	for i := uint32(0); i < maxID; i++ {
		nID := graph.NeuronID(i)
		meta := s.g.Meta[i]
		if meta.Type != "" || len(s.g.OutEdges(nID)) > 0 || len(s.g.InEdges(nID)) > 0 {
			nodes = append(nodes, MinimalNode{
				ID:     i,
				X:      meta.X,
				Y:      meta.Y,
				Z:      meta.Z,
				EmbedX: meta.EmbedX,
				EmbedY: meta.EmbedY,
				EmbedZ: meta.EmbedZ,
			})
		}
	}
	json.NewEncoder(w).Encode(nodes)
}

func (s *Server) handlePath(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	weightedStr := r.URL.Query().Get("weighted")

	from, err1 := strconv.ParseUint(fromStr, 10, 32)
	to, err2 := strconv.ParseUint(toStr, 10, 32)
	
	if err1 != nil || err2 != nil || uint32(from) > s.g.MaxNodeID() || uint32(to) > s.g.MaxNodeID() {
		http.Error(w, "invalid from/to ID", http.StatusBadRequest)
		return
	}

	var path []graph.NeuronID
	var found bool

	if weightedStr == "true" {
		path, found = query.ShortestPathWeighted(s.g, graph.NeuronID(from), graph.NeuronID(to))
	} else {
		path, found = query.ShortestPathHopCount(s.g, graph.NeuronID(from), graph.NeuronID(to))
	}

	if !found {
		http.Error(w, "no path found", http.StatusNotFound)
		return
	}

	layout := r.URL.Query().Get("layout")

	var resp GraphResponse
	for i, nID := range path {
		meta := s.g.Meta[nID]
		
		x, y, z := meta.X, meta.Y, meta.Z
		ex, ey, ez := meta.EmbedX, meta.EmbedY, meta.EmbedZ
		
		if layout == "logical" {
			x, y, z = meta.EmbedX, meta.EmbedY, meta.EmbedZ
			ex, ey, ez = meta.X, meta.Y, meta.Z
		}

		resp.Nodes = append(resp.Nodes, EnrichedNode{
			ID:     uint32(nID),
			Type:   meta.Type,
			X:      x,
			Y:      y,
			Z:      z,
			EmbedX: ex,
			EmbedY: ey,
			EmbedZ: ez,
		})
		if i > 0 {
			weight := uint16(1)
			for _, edge := range s.g.OutEdges(path[i-1]) {
				if edge.To == nID {
					weight = edge.Weight
					break
				}
			}
			resp.Edges = append(resp.Edges, EnrichedEdge{
				Source: uint32(path[i-1]),
				Target: uint32(nID),
				Weight: weight,
			})
		}
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleNeighborhood(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	depthStr := r.URL.Query().Get("depth")
	direction := r.URL.Query().Get("direction") // "up" or "down"

	id, err1 := strconv.ParseUint(idStr, 10, 32)
	depth, err2 := strconv.Atoi(depthStr)

	if err1 != nil || err2 != nil || depth < 1 || uint32(id) > s.g.MaxNodeID() {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}
	if direction != "up" && direction != "down" {
		direction = "down" // default
	}

	visited := query.Neighborhood(s.g, graph.NeuronID(id), depth, direction)
	
	if len(visited) > MaxRenderNodes {
		http.Error(w, "result too large for rendering (cap is 3000 nodes). please reduce depth.", http.StatusRequestEntityTooLarge)
		return
	}

	layout := r.URL.Query().Get("layout")

	var resp GraphResponse
	for nID := range visited {
		meta := s.g.Meta[nID]
		
		x, y, z := meta.X, meta.Y, meta.Z
		ex, ey, ez := meta.EmbedX, meta.EmbedY, meta.EmbedZ
		
		if layout == "logical" {
			x, y, z = meta.EmbedX, meta.EmbedY, meta.EmbedZ
			ex, ey, ez = meta.X, meta.Y, meta.Z
		}

		resp.Nodes = append(resp.Nodes, EnrichedNode{
			ID:     uint32(nID),
			Type:   meta.Type,
			X:      x,
			Y:      y,
			Z:      z,
			EmbedX: ex,
			EmbedY: ey,
			EmbedZ: ez,
		})
		
		for _, edge := range s.g.OutEdges(nID) {
			if visited[edge.To] {
				resp.Edges = append(resp.Edges, EnrichedEdge{
					Source: uint32(nID),
					Target: uint32(edge.To),
					Weight: edge.Weight,
				})
			}
		}
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleSubgraph(w http.ResponseWriter, r *http.Request) {
	nType := r.URL.Query().Get("type")
	if nType == "" {
		http.Error(w, "missing type parameter", http.StatusBadRequest)
		return
	}

	subgraph := query.ExtractSubgraphByType(s.g, nType)

	if len(subgraph.Nodes) > MaxRenderNodes {
		http.Error(w, "result too large for rendering (cap is 3000 nodes).", http.StatusRequestEntityTooLarge)
		return
	}

	layout := r.URL.Query().Get("layout")

	var resp GraphResponse
	for _, nID := range subgraph.Nodes {
		meta := s.g.Meta[nID]
		
		x, y, z := meta.X, meta.Y, meta.Z
		ex, ey, ez := meta.EmbedX, meta.EmbedY, meta.EmbedZ
		
		if layout == "logical" {
			x, y, z = meta.EmbedX, meta.EmbedY, meta.EmbedZ
			ex, ey, ez = meta.X, meta.Y, meta.Z
		}

		resp.Nodes = append(resp.Nodes, EnrichedNode{
			ID:     uint32(nID),
			Type:   meta.Type,
			X:      x,
			Y:      y,
			Z:      z,
			EmbedX: ex,
			EmbedY: ey,
			EmbedZ: ez,
		})
	}
	for _, edge := range subgraph.Edges {
		resp.Edges = append(resp.Edges, EnrichedEdge{
			Source: uint32(edge.From),
			Target: uint32(edge.To),
			Weight: edge.Weight,
		})
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleStatsDegree(w http.ResponseWriter, r *http.Request) {
	stats := query.GetBasicStats(s.g)
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleStatsTopConnected(w http.ResponseWriter, r *http.Request) {
	nStr := r.URL.Query().Get("n")
	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		n = 20
	}

	top := query.GetTopConnected(s.g, n)
	json.NewEncoder(w).Encode(top)
}
