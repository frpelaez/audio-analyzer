package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
)

const DecisionThreshold = 2000

func RunMatchCmd(args []string) {
	cmd := flag.NewFlagSet("match", flag.ExitOnError)
	threshold := cmd.Int("th", 100, "Threshold used for match decision ")
	debug := cmd.Bool("d", false, "debug only")

	cmd.Parse(args)

	if cmd.NArg() < 2 {
		fmt.Println("Error. Two .json files are needed for comparison")
		fmt.Println("Usage: audateci match <reference.json> <sample.json>")
		os.Exit(1)
	}

	refPath := cmd.Arg(0)
	samplePath := cmd.Arg(1)

	refData := loadFingerprint(refPath)
	sampleData := loadFingerprint(samplePath)

	fmt.Printf(
		"Comparing:\n   Reference: %s, (%d keypoints)\n   Sample:    %s, (%d keypoints)",
		refPath, len(refData.Points), samplePath, len(sampleData.Points))

	refIndex := make(map[int][]float64)
	for _, p := range refData.Points {
		freq := int(p.FreqHz)
		refIndex[freq] = append(refIndex[freq], p.TimeSec)
	}

	offsetHistogram := make(map[int]int)
	for _, pSample := range sampleData.Points {
		freq := int(pSample.FreqHz)
		timesRef, found := refIndex[freq]
		if found {
			for _, tRef := range timesRef {
				offset := tRef - pSample.TimeSec
				offsetBin := int(math.Round(offset * 10))
				offsetHistogram[offsetBin]++
			}
		}
	}

	bestOffsetBin := 0
	bestScore := 0
	for bin, score := range offsetHistogram {
		neighborsScore := score
		if bin > 0 {
			neighborsScore += offsetHistogram[bin-1]
		}
		if bin < len(offsetHistogram)-1 {
			neighborsScore += offsetHistogram[bin+1]
		}
		if neighborsScore > bestScore {
			bestOffsetBin = bin
			bestScore = neighborsScore
		}
	}

	predictedOffset := float64(bestOffsetBin) / 10.0

	fmt.Println("\nAnalysis results:")
	fmt.Printf("   Maximum score:    %d matches\n", bestScore)
	fmt.Printf("   Estimated offset: %.1f seconds\n", predictedOffset)

	conf := float64(bestScore)
	if conf > float64(*threshold) {
		fmt.Println("Results:")
		fmt.Println("   Match detected!")
		fmt.Printf("   The sample appears to be a fragment of the reference audio, starting at second %.1f", predictedOffset)
	} else {
		fmt.Println("   Sample did not match with the reference")
	}

	if *debug {
		exportHistogram(offsetHistogram)
	}
}

func loadFingerprint(path string) AudioFingerprint {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var data AudioFingerprint
	jsonerr := json.NewDecoder(f).Decode(&data)
	if jsonerr != nil {
		log.Fatal("Error decoding JSON: ", err)
	}

	return data
}

func exportHistogram(histogram map[int]int) {
	type Bin struct {
		Offset float64 `json:"offset"`
		Count  int     `json:"count"`
	}

	var data []Bin
	for k, v := range histogram {
		if v > 1 {
			data = append(data, Bin{Offset: float64(k) / 10.0, Count: v})
		}
	}

	sort.Slice(data, func(i, j int) bool { return data[i].Offset < data[j].Offset })

	file, _ := os.Create("debug/debug_hist.json")
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
	fmt.Println("\n(Debug) Histogram saved to 'debug/debug_hist.json'")
}
