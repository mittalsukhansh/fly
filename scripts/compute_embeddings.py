import pandas as pd
import networkx as nx
from node2vec import Node2Vec
import time
import argparse

def main():
    parser = argparse.ArgumentParser(description="Compute Node2Vec embeddings for graph")
    parser.add_argument("--edges", type=str, default="../edges.csv", help="Path to edges.csv")
    parser.add_argument("--nodes", type=str, default="../nodes.csv", help="Path to nodes.csv")
    parser.add_argument("--output", type=str, default="../embeddings.csv", help="Path to output embeddings.csv")
    args = parser.parse_args()
    
    start_time = time.time()
    
    print(f"Loading edges from {args.edges}...")
    edges_df = pd.read_csv(args.edges)
    
    print(f"Loaded {len(edges_df)} edges.")
    
    G = nx.from_pandas_edgelist(edges_df, source='from', target='to', edge_attr='weight', create_using=nx.Graph())
    
    try:
        nodes_df = pd.read_csv(args.nodes)
        for _, row in nodes_df.iterrows():
            if not G.has_node(row['id']):
                G.add_node(row['id'])
    except Exception as e:
        print(f"Could not load nodes.csv (or missing): {e}")

    print(f"Graph constructed: {G.number_of_nodes()} nodes, {G.number_of_edges()} edges.")
    
    print("Running Node2Vec...")
    node2vec = Node2Vec(G, dimensions=3, walk_length=15, num_walks=50, workers=1, quiet=False)
    
    print("Training model...")
    model = node2vec.fit(window=10, min_count=1, batch_words=4)
    
    print("Extracting embeddings...")
    embeddings = []
    for node in G.nodes():
        vec = model.wv[str(node)]
        embeddings.append({
            'node_id': int(node),
            'embed_x': vec[0],
            'embed_y': vec[1],
            'embed_z': vec[2]
        })
        
    out_df = pd.DataFrame(embeddings)
    out_df.to_csv(args.output, index=False)
    
    duration = time.time() - start_time
    print(f"Finished generating embeddings in {duration:.2f} seconds.")
    print(f"Saved to {args.output}")

if __name__ == "__main__":
    main()
