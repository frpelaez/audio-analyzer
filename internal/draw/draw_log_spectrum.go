package draw

import (
	"fmt"
	"math"
	"strings"
)

func DrawLogSpectrum(magnitudes []float64, sampleRate int, numBars int) {
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

		minDB := -70.0
		maxDB := 0.0

		normalzied := (dbVal - minDB) / (maxDB - minDB)
		normalzied = max(0, min(1, normalzied))

		barLen := int(normalzied * 70)
		barBuilder := strings.Builder{}
		// for i := range barLen {
		// 	if i == 0 {
		// 		barBuilder.WriteString("|█|")
		// 	}
		// 	barBuilder.WriteString("█|")
		// }

		for range barLen {
			barBuilder.WriteString("█")
		}

		freqLabel := fmt.Sprintf("%0.fHz", highF)
		fmt.Printf("\n%8s | %s (%.1f dB)", freqLabel, barBuilder.String(), dbVal)
	}
}
