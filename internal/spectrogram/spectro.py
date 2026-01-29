import argparse
import sys

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd


def spectrogram(csv_file: str, output_file=None):
    print(f"Loading data from: {csv_file}")

    try:
        df = pd.read_csv(csv_file)
    except FileNotFoundError:
        print("Error: csv file not found")
        sys.exit(1)

    times = df.iloc[:, 0].values
    data = df.iloc[:, 1:].values.T

    time_max = times.max()

    last_freq = df.columns[-1]
    try:
        max_freq = float(last_freq.replace("Hz", ""))
    except ValueError:
        max_freq = 22050

    num_bins = data.shape[0]
    freq_step = max_freq / (num_bins - 1)

    print(f"Frequency resolution: {freq_step:.2f} Hz per bin")

    data_log = data[1:, :]

    freqs = np.linspace(freq_step, max_freq, data_log.shape[0])
    print(f"Duration: {time_max:.2f}s")
    print(f"Data dimensions: {data_log.shape}")
    print(f"Frequency range: {freqs.min():.1f}Hz - {freqs.max():.1f}Hz")

    fig, ax = plt.subplots(figsize=(12, 6), dpi=100)

    img = ax.pcolormesh(times, freqs, data_log, cmap="inferno", shading="auto")

    ax.set_yscale("log")

    plt.colorbar(img, label="Magnitude (dB)", ax=ax)
    ax.set_title(f"Log-spectrogram: {csv_file}")
    ax.set_xlabel("Time (s)")
    ax.set_ylabel("Frequency (Hz)")

    ax.set_xlim(times.min(), times.max())
    ax.set_ylim(max(20, freqs.min()), max_freq)

    plt.tight_layout()

    if output_file:
        print(f"Saving image to: {output_file}")
        plt.savefig(output_file, dpi=100, bbox_inches="tight")
        print("Image saved successfully")
    else:
        print("Generating image...")
        plt.show()


def main():
    parser = argparse.ArgumentParser(description="Spectrogram visualizer from CSV")
    parser.add_argument(
        "file",
        type=str,
        help="Path to .csv file containing the audio time vs. frequency data",
    )
    parser.add_argument("--save", type=str, help="Target path")

    args = parser.parse_args()
    spectrogram(args.file, output_file=args.save)


if __name__ == "__main__":
    main()
