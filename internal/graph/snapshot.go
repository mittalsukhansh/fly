package graph

import (
	"encoding/gob"
	"fmt"
	"os"
)

// SaveSnapshot serializes the Graph to a binary file using gob.
func (g *Graph) SaveSnapshot(path string) error {
	fmt.Printf("Saving graph snapshot to %s...\n", path)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(g); err != nil {
		return fmt.Errorf("failed to encode graph: %w", err)
	}

	fmt.Println("Snapshot saved.")
	return nil
}

// LoadSnapshot deserializes the Graph from a binary file.
func LoadSnapshot(path string) (*Graph, error) {
	fmt.Printf("Loading graph snapshot from %s...\n", path)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var g Graph
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&g); err != nil {
		return nil, fmt.Errorf("failed to decode graph: %w", err)
	}

	fmt.Println("Snapshot loaded successfully.")
	return &g, nil
}
