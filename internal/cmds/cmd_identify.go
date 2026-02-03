package cmd

import (
	"audateci/internal/signal"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
)

type IndexEntry struct {
	SongName string
	TimeSec  float64
}

type FingerprintFile struct {
	Filename string            `json:"filename"`
	Points   []signal.KeyPoint `json:"points"`
}

func RunIdentifyCmd(args []string) {
	cmd := flag.NewFlagSet("identify", flag.ExitOnError)
	winsize := cmd.Int("winsize", 2048, "Size of the FFT window (must match the one sued to create the fingerprints)")
	threshold := cmd.Int("th", 100, "Threshold for the number of matches")

	cmd.Parse(args)

	if cmd.NArg() < 2 {
		fmt.Println("Usage: audateci identify <directory-with-fingerprints> <audio-fragment.wav>")
		os.Exit(1)
	}

	dbFolder := cmd.Arg(0)
	fragmentPath := cmd.Arg(1)

	fmt.Printf("Indexing db directory: '%s'\n", dbFolder)

	dbIndex := make(map[int][]IndexEntry)

	files, err := filepath.Glob(filepath.Join(dbFolder, "*.json"))
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		log.Fatalf("No fingerprint json files found in '%s'\n", dbFolder)
	}

	for _, file := range files {
		f, _ := os.Open(file)
		var fingerprint FingerprintFile
		json.NewDecoder(f).Decode(&fingerprint)
		f.Close()

		for _, p := range fingerprint.Points {
			freq := int(p.FreqHz)
			entry := IndexEntry{SongName: fingerprint.Filename, TimeSec: p.TimeSec}
			dbIndex[freq] = append(dbIndex[freq], entry)
		}
	}
	fmt.Printf("%d songs indexed from local database '%s'\n", len(files), dbFolder)

	fmt.Printf("Analyzing fragment '%s'\n", fragmentPath)

	fragmentKeypoints := signal.GetKeypointsFromFile(fragmentPath, *winsize)

	fmt.Printf(
		"Successfully generated fingerprint for input fragment (%d keypoints)\n",
		len(fragmentKeypoints),
	)

	scores := make(map[string]map[int]int)
	for _, pFrag := range fragmentKeypoints {
		freq := int(pFrag.FreqHz)
		entries, found := dbIndex[freq]
		if found {
			for _, entry := range entries {
				offset := entry.TimeSec - pFrag.TimeSec
				offsetBin := int(math.Round(offset * 10))

				if scores[entry.SongName] == nil {
					scores[entry.SongName] = make(map[int]int)
				}
				scores[entry.SongName][offsetBin]++
			}
		}
	}

	bestSong := ""
	bestOffset := 0.0
	bestScore := 0
	for song, offsetMap := range scores {
		for bin, count := range offsetMap {
			neighborsScore := count
			if val, ok := offsetMap[bin-1]; ok {
				neighborsScore += val
			}
			if val, ok := offsetMap[bin+1]; ok {
				neighborsScore += val
			}
			if neighborsScore > bestScore {
				bestSong = song
				bestOffset = float64(bin) / 10.0
				bestScore = neighborsScore
			}
		}
	}

	fmt.Println("Results:")
	if bestScore > *threshold {
		fmt.Printf("   Song:   %s\n", bestSong)
		fmt.Printf("   Offset: %.1fs\n", bestOffset)
		fmt.Printf("   Score:  %d matches\n", bestScore)
	} else {
		fmt.Printf("   No clear matches with decision threshold %d\n", *threshold)
	}
}
