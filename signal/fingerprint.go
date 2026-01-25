package signal

import (
	"math"
)

type FreqRange struct {
	Min float64
	Max float64
}

type KeyPoint struct {
	TimeSec float64 `json:"t"`
	FreqHz  float64 `json:"f"`
	MagDB   float64 `json:"m"`
}

var Bands = []FreqRange{
	{Min: 40, Max: 300},
	{Min: 300, Max: 2000},
	{Min: 2000, Max: 5000},
	{Min: 5000, Max: 10000},
}

func GetFingerprintPoints(magnitudes []float64, sampleRate int, windowSize int, currentTime float64) []KeyPoint {
	var points []KeyPoint

	silence := -50.0

	for _, band := range Bands {
		maxMag := -999.0
		maxIdx := -1

		startBin := int(band.Min * float64(windowSize) / float64(sampleRate))
		endBin := int(band.Max * float64(windowSize) / float64(sampleRate))
		startBin = max(0, startBin)
		endBin = min(endBin, len(magnitudes)-1)

		for k := startBin; k <= endBin; k++ {
			magDB :=20.0 * math.Log10(magnitudes[k])
			if magDB > maxMag {
				maxMag = magDB
				maxIdx = k
			}
		}

		if maxIdx != -1 && maxMag > silence {
			freq := float64(maxIdx) * float64(sampleRate) / float64(windowSize)
			points = append(points, KeyPoint{
				TimeSec: math.Round(currentTime*1000) / 1000,
				FreqHz:  math.Round(freq),
				MagDB:   math.Round(maxMag*100) / 100,
			})
		}
	}

	return points
}
