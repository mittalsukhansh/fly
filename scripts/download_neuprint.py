"""
neuPrint full hemibrain dataset downloader.
Fetches neurons + connections and saves as nodes.csv / edges.csv.
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

def cypher(query, params=None):
    payload = {"cypher": query, "dataset": DATASET}
    if params:
        payload["parameters"] = params
    r = requests.post(f"{BASE_URL}/api/custom/custom", json=payload, headers=HEADERS, timeout=120)
    r.raise_for_status()
    return r.json()

def fetch_neurons(out_path="nodes.csv"):
    print("Fetching neurons...")
    # Get count first
    count_q = "MATCH (n:Neuron) RETURN count(n) as c"
    total = cypher(count_q)["data"][0][0]
    print(f"  Total neurons: {total}")

    BATCH = 5000
    written = 0
    with open(out_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.writer(f)
        writer.writerow(["id", "type", "x", "y", "z"])

        for skip in range(0, total, BATCH):
            q = f"""
            MATCH (n:Neuron)
            RETURN n.bodyId, n.type,
                   n.somaLocation.x, n.somaLocation.y, n.somaLocation.z
            ORDER BY n.bodyId
            SKIP {skip} LIMIT {BATCH}
            """
            result = cypher(q)
            rows = result.get("data", [])
            for row in rows:
                body_id = row[0]
                ntype   = row[1] or ""
                x       = row[2] or 0
                y       = row[3] or 0
                z       = row[4] or 0
                writer.writerow([body_id, ntype, x, y, z])
            written += len(rows)
            print(f"  Neurons: {written}/{total}", end="\r", flush=True)
            if len(rows) < BATCH:
                break
            time.sleep(0.2)

    print(f"\n  Saved {written} neurons -> {out_path}")
    return written

def fetch_connections(out_path="edges.csv"):
    print("Fetching connections...")
    count_q = "MATCH (:Neuron)-[c:ConnectsTo]->(:Neuron) RETURN count(c) as c"
    total = cypher(count_q)["data"][0][0]
    print(f"  Total connections: {total}")

    BATCH = 10000
    written = 0
    with open(out_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.writer(f)
        writer.writerow(["pre_id", "post_id", "weight"])

        for skip in range(0, total, BATCH):
            q = f"""
            MATCH (a:Neuron)-[c:ConnectsTo]->(b:Neuron)
            RETURN a.bodyId, b.bodyId, c.weight
            ORDER BY a.bodyId
            SKIP {skip} LIMIT {BATCH}
            """
            result = cypher(q)
            rows = result.get("data", [])
            for row in rows:
                writer.writerow([row[0], row[1], row[2] or 1])
            written += len(rows)
            print(f"  Connections: {written}/{total}", end="\r", flush=True)
            if len(rows) < BATCH:
                break
            time.sleep(0.2)

    print(f"\n  Saved {written} connections -> {out_path}")
    return written

if __name__ == "__main__":
    out_dir = sys.argv[1] if len(sys.argv) > 1 else "."

    # Quick auth check
    print("Verifying token...")
    try:
        info = cypher("RETURN 1")
        print("  Auth OK")
    except requests.HTTPError as e:
        print(f"  Auth FAILED: {e.response.status_code} {e.response.text[:200]}")
        sys.exit(1)

    fetch_neurons(f"{out_dir}/nodes.csv")
    fetch_connections(f"{out_dir}/edges.csv")
    print("\nDone! Now restart the Go backend to ingest the new data.")
