import json
import sys

import matplotlib.pyplot as plt


def plot_histogram(json_file):
    try:
        with open(json_file) as f:
            data = json.load(f)
    except FileNotFoundError:
        print("Unnable to find file 'debug/debug_hist.py'")
        sys.exit(1)

    offsets = [d["offset"] for d in data]
    counts = [d["count"] for d in data]
    smooth_counts = [0 for _ in range(len(counts))]
    smooth_counts[0] = counts[0]
    for i in range(1, len(smooth_counts) - 1):
        smooth_counts[i] = sum(counts[i-1:i+2])
    smooth_counts[-1] = counts[-1]

    if not offsets:
        print("Did not find any significative matches")
        sys.exit(1)

    plt.figure(figsize=(10, 6))

    plt.bar(offsets, smooth_counts, width=0.1, color="teal", alpha=0.7)

    max_count = max(counts)
    max_idx = counts.index(max_count)
    best_offset = offsets[max_idx]

    plt.title(f"Alignment histogram\nDetected peak at {best_offset}s")
    plt.xlabel("Offset (seconds)")
    plt.ylabel("Number of matching points")
    plt.grid(True, alpha=0.3)

    plt.annotate(
        "Match",
        xy=(best_offset, max_count),
        xytext=(best_offset, max_count + 0.15 * max_count),
        arrowprops=dict(facecolor="red", shrink=0.05),
        horizontalalignment="center",
    )

    plt.tight_layout()
    plt.show()


if __name__ == "__main__":
    plot_histogram("./debug/debug_hist.json")
