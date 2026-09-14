"""
Download all synaptic connections from neuPrint hemibrain:v1.1
Strategy: paginate using bodyId ranges (min/max) instead of SKIP,
which neuPrint rejects at large offsets.
"""
import requests
import csv
import sys
import time

TOKEN = "e6415fa047b749e0d140f4a1f1e769bae1905632c5dd9c61c87d9f3e3023d5ac"
BASE_URL = "https://neuprint.janelia.org"
DATASET = "hemibrain:v1.1"
HEADERS = {
    "Authorization": f"Bearer {TOKEN}",
    "Content-Type": "application/json",
}

def cypher(query):
    r = requests.post(f"{BASE_URL}/api/custom/custom",
                      json={"cypher": query, "dataset": DATASET},
                      headers=HEADERS, timeout=120)
    if not r.ok:
        raise RuntimeError(f"HTTP {r.status_code}: {r.text[:300]}")
    return r.json().get("data", [])

def load_neuron_ids(path):
    ids = []
    with open(path, "r", encoding="utf-8") as f:
        reader = csv.reader(f)
        next(reader)
        for row in reader:
            if row:
                ids.append(int(row[0]))
    return sorted(ids)

if __name__ == "__main__":
    out_dir = sys.argv[1] if len(sys.argv) > 1 else "."

    print("Loading neuron IDs from nodes.csv...")
    ids = load_neuron_ids(f"{out_dir}/nodes.csv")
    print(f"  {len(ids):,} neurons | ID range: {ids[0]} – {ids[-1]}")

    # Divide the ID space into chunks and query connections per chunk
    CHUNK = 2000  # neurons per chunk
    total_chunks = (len(ids) + CHUNK - 1) // CHUNK
    written = 0
    errors = 0

    out_path = f"{out_dir}/edges.csv"
    with open(out_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.writer(f)
        writer.writerow(["pre_id", "post_id", "weight"])

        for ci, start in enumerate(range(0, len(ids), CHUNK)):
            chunk = ids[start : start + CHUNK]
            min_id = chunk[0]
            max_id = chunk[-1]

            q = f"""
            MATCH (a:Neuron)-[c:ConnectsTo]->(b:Neuron)
            WHERE a.bodyId >= {min_id} AND a.bodyId <= {max_id}
            RETURN a.bodyId, b.bodyId, c.weight
            """

            retries = 3
            for attempt in range(retries):
                try:
                    rows = cypher(q)
                    for row in rows:
                        writer.writerow([row[0], row[1], row[2] or 1])
                    written += len(rows)
                    break
                except Exception as e:
                    if attempt == retries - 1:
                        errors += 1
                        print(f"\n  Chunk {ci} failed after {retries} retries: {e}")
                    else:
                        time.sleep(3)

            if ci % 20 == 0:
                pct = (ci / total_chunks) * 100
                print(f"  [{pct:5.1f}%] chunk {ci+1}/{total_chunks} — {written:,} edges so far", end="\r", flush=True)

            time.sleep(0.1)

    print(f"\n  Done. {written:,} connections saved -> {out_path}  ({errors} chunk errors)")
    print("\nNext: delete connectome.bin and restart the Go backend.")
