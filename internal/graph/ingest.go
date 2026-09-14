package graph

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

// IngestData streams the raw node and edge CSV files to construct the graph.
func IngestData(nodesPath, edgesPath, embeddingsPath string) (*Graph, error) {
	fmt.Println("Starting Ingest (Phase 1 & 2)...")

	// 1. First pass over edges: count degrees and find maxNodeID
	fmt.Println("  Pass 1: Scanning edges to compute degrees...")
	builder := NewGraphBuilder(166700) // Initial capacity estimate

	edgeFile, err := os.Open(edgesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open edges file: %w", err)
	}

	reader := csv.NewReader(bufio.NewReader(edgeFile))
	// Assume header: pre_id, post_id, weight
	_, err = reader.Read()
	if err != nil {
		edgeFile.Close()
		return nil, fmt.Errorf("failed to read edge header: %w", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			edgeFile.Close()
			return nil, fmt.Errorf("error reading edge: %w", err)
		}

		fromID, err := strconv.ParseUint(record[0], 10, 32)
		if err != nil {
			continue
		}
		toID, err := strconv.ParseUint(record[1], 10, 32)
		if err != nil {
			continue
		}
		weight, err := strconv.ParseUint(record[2], 10, 16)
		if err != nil {
			weight = 1
		}

		builder.AddEdgePass1(NeuronID(fromID), NeuronID(toID), uint16(weight))
	}
	edgeFile.Close()

	// 2. Read nodes (metadata)
	fmt.Printf("  Reading nodes (Max Node ID: %d)...\n", builder.maxNodeID)
	meta := make([]NeuronMeta, builder.maxNodeID+1)

	nodeFile, err := os.Open(nodesPath)
	if err == nil {
		nReader := csv.NewReader(bufio.NewReader(nodeFile))
		// Assume header: id, type, x, y, z
		_, _ = nReader.Read()
		for {
			record, err := nReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}
			id, err := strconv.ParseUint(record[0], 10, 32)
			if err != nil || uint32(id) > builder.maxNodeID {
				continue
			}
			
			nodeType := ""
			if len(record) > 1 {
				nodeType = record[1]
			}
			
			var x, y, z float64
			if len(record) > 4 {
				x, _ = strconv.ParseFloat(record[2], 32)
				y, _ = strconv.ParseFloat(record[3], 32)
				z, _ = strconv.ParseFloat(record[4], 32)
			}
			
			meta[id] = NeuronMeta{
				Type: nodeType,
				X:    float32(x),
				Y:    float32(y),
				Z:    float32(z),
			}
		}
		nodeFile.Close()
	} else {
		fmt.Printf("  Warning: nodes file not found, skipping metadata (%v)\n", err)
	}

	// 2b. Read embeddings
	fmt.Printf("  Reading embeddings from %s...\n", embeddingsPath)
	embedFile, err := os.Open(embeddingsPath)
	if err == nil {
		eReader := csv.NewReader(bufio.NewReader(embedFile))
		// Assume header: node_id, embed_x, embed_y, embed_z
		_, _ = eReader.Read()
		for {
			record, err := eReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}
			id, err := strconv.ParseUint(record[0], 10, 32)
			if err != nil || uint32(id) > builder.maxNodeID {
				continue
			}
			
			if len(record) >= 4 {
				ex, _ := strconv.ParseFloat(record[1], 32)
				ey, _ := strconv.ParseFloat(record[2], 32)
				ez, _ := strconv.ParseFloat(record[3], 32)
				
				meta[id].EmbedX = float32(ex)
				meta[id].EmbedY = float32(ey)
				meta[id].EmbedZ = float32(ez)
			}
		}
		embedFile.Close()
	} else {
		fmt.Printf("  Warning: embeddings file not found, logical layout will be empty (%v)\n", err)
	}

	// 3. Allocate CSR
	fmt.Println("  Allocating CSR structures...")
	builder.Allocate(meta)

	// 4. Second pass over edges: populate CSR arrays
	fmt.Println("  Pass 2: Populating CSR edges...")
	edgeFile2, err := os.Open(edgesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open edges file for pass 2: %w", err)
	}
	defer edgeFile2.Close()

	reader2 := csv.NewReader(bufio.NewReader(edgeFile2))
	_, _ = reader2.Read() // skip header

	for {
		record, err := reader2.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		fromID, _ := strconv.ParseUint(record[0], 10, 32)
		toID, _ := strconv.ParseUint(record[1], 10, 32)
		weight, err := strconv.ParseUint(record[2], 10, 16)
		if err != nil {
			weight = 1
		}

		builder.AddEdgePass2(NeuronID(fromID), NeuronID(toID), uint16(weight))
	}

	fmt.Println("Ingest complete.")
	return builder.Build(), nil
}
