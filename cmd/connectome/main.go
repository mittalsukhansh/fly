package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fly/connectomegraph/internal/api"
	"github.com/fly/connectomegraph/internal/graph"
)

func main() {
	snapshotPath := flag.String("snapshot", "connectome.bin", "Path to graph snapshot file")
	nodesCSV := flag.String("nodes", "nodes.csv", "Path to nodes CSV (only used if snapshot not found)")
	edgesCSV := flag.String("edges", "edges.csv", "Path to edges CSV (only used if snapshot not found)")
	embeddingsCSV := flag.String("embeddings", "embeddings.csv", "Path to logical embeddings CSV (only used if snapshot not found)")
	addr := flag.String("addr", ":8080", "HTTP server address")
	flag.Parse()

	var g *graph.Graph
	var err error

	if _, statErr := os.Stat(*snapshotPath); statErr == nil {
		fmt.Println("Snapshot found, loading...")
		g, err = graph.LoadSnapshot(*snapshotPath)
		if err != nil {
			log.Fatalf("Failed to load snapshot: %v", err)
		}
	} else {
		fmt.Println("Snapshot not found, starting ingest...")
		g, err = graph.IngestData(*nodesCSV, *edgesCSV, *embeddingsCSV)
		if err != nil {
			log.Fatalf("Failed to ingest data: %v", err)
		}
		
		fmt.Println("Saving snapshot for next run...")
		if err := g.SaveSnapshot(*snapshotPath); err != nil {
			log.Printf("Warning: failed to save snapshot: %v", err)
		}
	}

	g.PrintStats()

	fmt.Printf("Starting HTTP API on %s...\n", *addr)
	server := api.NewServer(g)
	if err := server.Start(*addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
