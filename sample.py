import requests
import time
import matplotlib.pyplot as plt
import pandas as pd
import numpy as np

GLOBAL_LB_URL = "http://34.120.142.19/"  
MAIN_SERVER_URL = "http://35.200.253.69:5000/"         
EDGE_NODES = {
    "us-central1": "https://cdn-node-1-774700561092.us-central1.run.app/",
    "southamerica-east1": "https://cdn-node-2-774700561092.southamerica-east1.run.app/",
    "europe-west2": "https://cdn-node-3-774700561092.europe-west2.run.app/",
    "africa-south1": "https://cdn-node-4-774700561092.africa-south1.run.app/",
    "asia-south2": "https://cdn-node-5-774700561092.asia-south2.run.app/",
    "asia-northeast1": "https://cdn-node-6-774700561092.asia-northeast1.run.app/",
    "australia-southeast1": "https://cdn-node-7-774700561092.australia-southeast1.run.app/"
}

TEST_KEY = "test123"
TEST_PAYLOAD = {"value": "VGhpcyBpcyBhIHNhbXBsZSBjb250ZW50IHRvIHRlc3QgdGhlIENETi4="}
REPEAT = 20

def post_data(url):
    return requests.post(url + TEST_KEY, json=TEST_PAYLOAD)

def delete_data(url):
    return requests.delete(url + TEST_KEY)

def get_data(url):
    t0 = time.time()
    res = requests.get(url + TEST_KEY)
    t1 = time.time()
    latency = (t1 - t0) * 1000  
    return latency, res.status_code


def run_tests1(label, url):
    print(f"\n--- {label} ---")
    
    delete_data(url)
    
    print("Cold request...")
    post_data(url)  
    delete_data(url)
    post_data(url)  
    latency_cold, _ = get_data(url)
    
    print("Warm requests...")
    warm_latencies = []
    for _ in range(REPEAT):
        latency, _ = get_data(url)
        warm_latencies.append(latency)

    return latency_cold, warm_latencies

def run_tests2(label, url):
    print(f"\n--- {label} ---")
    
    print("Cold request...")
    latency_cold, _ = get_data(url)
    
    print("Warm requests...")
    warm_latencies = []
    for _ in range(REPEAT):
        latency, _ = get_data(url)
        warm_latencies.append(latency)
    
    return latency_cold, warm_latencies

results = {}

cold, warm = run_tests1("Main Server (Direct)", MAIN_SERVER_URL)
results["Main Server"] = {"cold": cold, "warm": warm}

cold, warm = run_tests2("Global Load Balancer", GLOBAL_LB_URL)
results["Global LB"] = {"cold": cold, "warm": warm}

for region, url in EDGE_NODES.items():
    cold, warm = run_tests2(f"Node: {region}", url)
    results[region] = {"cold": cold, "warm": warm}

labels = list(results.keys())
cold_latencies = [results[k]["cold"] for k in labels]
cleaned_latencies = sorted(results[k]["warm"][1:-1] for k in labels)  
avg_warm_latencies = [sum(cleaned_latencies[k])/len(cleaned_latencies[k]) for k in range(len(labels))]

plt.figure(figsize=(12, 6))
x = range(len(labels))

plt.bar(x, cold_latencies, width=0.4, label='Cold (Uncached)', align='center', alpha=0.7)
plt.bar([p + 0.4 for p in x], avg_warm_latencies, width=0.4, label='Warm (Cached)', align='center', alpha=0.7)
plt.xticks([p + 0.2 for p in x], labels, rotation=30)
plt.ylabel("Latency (ms)")
plt.title("CDN Performance Evaluation")
plt.legend()
plt.grid(True)
plt.tight_layout()
plt.show()

table_data = []
for label in labels:
    cold = round(results[label]["cold"], 2)
    warm_avg = np.mean(results[label]["warm"])
    warm_std = np.std(results[label]["warm"])
    table_data.append([
        label,
        f"{cold} ms",
        f"{round(warm_avg, 2)} ± {round(warm_std, 2)} ms"
    ])

df = pd.DataFrame(table_data, columns=["Endpoint", "Cold Latency", "Warm Latency (Avg ± Std Dev)"])
print("\nFinal Results Summary Table:")
display(df)