import argparse
import json
import sys

import matplotlib.pyplot as plt
import scipy.io.wavfile as wav

BANDS = [
    (40, 300),  # Bajos
    (300, 2000),  # Medios
    (2000, 5000),  # Medios-Altos
    (5000, 10000),  # Agudos
]


def verify_fingerprint(wav_file, json_file):
    print(f"Audio: {wav_file}")
    print(f"Fingerprint: {json_file}")

    try:
        sample_rate, data = wav.read(wav_file)
        if len(data.shape) > 1:
            data = data[:, 0]
    except FileNotFoundError:
        print("Error: Audio file not found.")
        sys.exit(1)

    try:
        with open(json_file, "r") as f:
            fingerprint = json.load(f)
    except FileNotFoundError:
        print("Error: JSOn file not found.")
        sys.exit(1)

    points = fingerprint["points"]
    t_coords = [p["t"] for p in points]
    f_coords = [p["f"] for p in points]

    print(f"Loaded points: {len(points)}")

    plt.figure(figsize=(14, 8))

    Pxx, freqs, bins, im = plt.specgram(
        data, NFFT=2048, Fs=sample_rate, noverlap=1024, cmap="gray_r"
    )

    plt.scatter(
        t_coords, f_coords, color="red", s=25, marker="*", label="Detected peaks"
    )

    for b_min, b_max in BANDS:
        plt.axhline(y=b_max, color="blue", linestyle="--", linewidth=0.8, alpha=0.5)

    plt.title(f"Constelation map: {wav_file}")
    plt.ylabel("Frequency (Hz)")
    plt.xlabel("Time (s)")
    plt.legend(loc="upper right")

    plt.ylim(0, 11000)
    plt.xlim(0, fingerprint["duration"])

    plt.tight_layout()
    plt.show()


def main():
    parser = argparse.ArgumentParser(description="Verifier for audio fingerprints")
    parser.add_argument("wav", type=str, help="Original audio file .wav")
    parser.add_argument("json", type=str, help="Fingerprint file (.json)")

    args = parser.parse_args()
    verify_fingerprint(args.wav, args.json)


if __name__ == "__main__":
    main()
