package export

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/fly/connectomegraph/internal/graph"
	"github.com/fly/connectomegraph/internal/query"
)

type GraphML struct {
	XMLName xml.Name `xml:"graphml"`
	Xmlns   string   `xml:"xmlns,attr"`
	Key     []Key    `xml:"key"`
	Graph   Graph    `xml:"graph"`
}

type Key struct {
	ID   string `xml:"id,attr"`
	For  string `xml:"for,attr"`
	Name string `xml:"attr.name,attr"`
	Type string `xml:"attr.type,attr"`
}

type Graph struct {
	ID         string `xml:"id,attr"`
	EdgeDefault string `xml:"edgedefault,attr"`
	Nodes      []Node `xml:"node"`
	Edges      []Edge `xml:"edge"`
}

type Node struct {
	ID   string `xml:"id,attr"`
	Data []Data `xml:"data"`
}

type Edge struct {
	Source string `xml:"source,attr"`
	Target string `xml:"target,attr"`
	Data   []Data `xml:"data"`
}

type Data struct {
	Key   string `xml:"key,attr"`
	Value string `xml:",chardata"`
}

// ExportGraphML saves a subgraph to GraphML format.
func ExportGraphML(subgraph query.SubgraphResponse, meta []graph.NeuronMeta, path string) error {
	graphML := GraphML{
		Xmlns: "http://graphml.graphdrawing.org/xmlns",
		Key: []Key{
			{ID: "type", For: "node", Name: "type", Type: "string"},
			{ID: "weight", For: "edge", Name: "weight", Type: "int"},
		},
		Graph: Graph{
			ID:         "G",
			EdgeDefault: "directed",
		},
	}

	for _, nID := range subgraph.Nodes {
		m := meta[nID]
		graphML.Graph.Nodes = append(graphML.Graph.Nodes, Node{
			ID: fmt.Sprintf("%d", nID),
			Data: []Data{
				{Key: "type", Value: m.Type},
			},
		})
	}

	for _, edge := range subgraph.Edges {
		graphML.Graph.Edges = append(graphML.Graph.Edges, Edge{
			Source: fmt.Sprintf("%d", edge.From),
			Target: fmt.Sprintf("%d", edge.To),
			Data: []Data{
				{Key: "weight", Value: fmt.Sprintf("%d", edge.Weight)},
			},
		})
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString(xml.Header)
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	return encoder.Encode(graphML)
}
