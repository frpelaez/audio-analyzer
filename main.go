package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "listen":
		runListenCmd(os.Args[2:])
	case "analyze":
		runAnalyzeCmd(os.Args[2:])
	case "spectro":
		runSpectroCmd(os.Args[2:])
	case "-h", "--help", "help":
		printHelp()
	default:
		fmt.Printf("Unkown command: '%s'\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Usage: audateci <command> [options] <audio_file.wav>")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  listen    Visualize the frequencies contained in the audio file")
	fmt.Println("  analyze   Analyze the audio file and export data to CSV/Bin (wip)")
	fmt.Println("  spectro   Compute spectrogram from audio file and export to png")
	fmt.Println("\nType audateci <command> -h for specific help")
}

func drawLogSpectrum(magnitudes []float64, sampleRate int, numBars int) {
	fmt.Println("\n--- Logarithmic Frequency Spectrum ---")

	minFreq := 20.0
	maxFreq := 20_000.0

	logScale := math.Log(maxFreq / minFreq)

	for i := range numBars {
		lowF := minFreq * math.Exp(logScale*float64(i)/float64(numBars))
		highF := minFreq * math.Exp(logScale*float64(i+1)/float64(numBars))

		idxStart := int(lowF * float64(len(magnitudes)) / float64(sampleRate))
		idxEnd := int(highF * float64(len(magnitudes)) / float64(sampleRate))

		idxStart = max(idxStart, 0)
		idxEnd = min(idxEnd, len(magnitudes)-1)
		if idxStart >= idxEnd {
			idxEnd = idxStart + 1
		}

		maxMagInBin := 0.0
		for j := idxStart; j < idxEnd; j++ {
			if magnitudes[j] > maxMagInBin {
				maxMagInBin = magnitudes[j]
			}
		}

		epsilon := 1e-9
		maxMagInBin = max(maxMagInBin, epsilon)
		dbVal := 20 * math.Log10(maxMagInBin)

		minDB := -60.0
		maxDB := 0.0

		normalzied := (dbVal - minDB) / (maxDB - minDB)
		normalzied = max(0, min(1, normalzied))

		barLen := int(normalzied * 40)
		barBuilder := strings.Builder{}
		for i := range barLen {
			if i == 0 {
				barBuilder.WriteString("|█|")
			}
			barBuilder.WriteString("█|")
		}

		freqLabel := fmt.Sprintf("%0.fHz", highF)
		fmt.Printf("\n%8s | %s (%.1f dB)", freqLabel, barBuilder.String(), dbVal)
	}
}
