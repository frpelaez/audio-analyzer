package cmd

import (
	"audateci/internal/signal"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type IndexEntry struct {
	SongName string
	TimeSec  float64
}

type FingerprintFile struct {
	Filename string            `json:"filename"`
	Points   []signal.KeyPoint `json:"points"`
}

type MatchResult struct {
	QueryFile   string
	BestMatch   string
	Offset      float64
	Score       int
	TotalPoints int
	Confidence  float64
	ProcessTime time.Duration
}

var windowSize = 2048

const ConfidenceThreshold = 3.0

func RunIdentifyCmd(args []string) {
	cmd := flag.NewFlagSet("identify", flag.ExitOnError)
	cmd.IntVar(&windowSize, "winsize", 2048, "Size of the FFT window (must match the one sued to create the fingerprints)")
	outputFile := cmd.String("csv", "reports/test_results.csv", "Name for the report file (only for batch mode)")

	cmd.Parse(args)

	if cmd.NArg() < 2 {
		fmt.Println("Usage: audateci identify <directory-with-fingerprints> <audio-fragment.wav>")
		os.Exit(1)
	}

	dbFolder := cmd.Arg(0)
	inputPath := cmd.Arg(1)

	fmt.Printf("Indexing db directory: '%s'\n", dbFolder)
	dbIndex := loadDatabase(dbFolder)

	info, err := os.Stat(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	if info.IsDir() {
		_runBatchMode(inputPath, dbIndex, *outputFile)
	} else {
		runSingleMode(inputPath, dbIndex)
	}
}

func loadDatabase(path string) map[int][]IndexEntry {
	invertedIndex := make(map[int][]IndexEntry)
	files, _ := filepath.Glob(filepath.Join(path, "*.json"))

	for _, file := range files {
		f, _ := os.Open(file)
		var fp FingerprintFile
		json.NewDecoder(f).Decode(&fp)
		f.Close()

		for _, p := range fp.Points {
			freq := int(p.FreqHz)
			invertedIndex[freq] = append(invertedIndex[freq], IndexEntry{
				SongName: fp.Filename,
				TimeSec:  p.TimeSec,
			})
		}
	}

	return invertedIndex
}

func identifyAudio(path string, index map[int][]IndexEntry) (MatchResult, error) {
	startTime := time.Now()

	data, err := signal.ReadWavToFloats(path)
	if err != nil {
		return MatchResult{}, err
	}
	samples := data.Channels[0]

	var queryPoints []signal.KeyPoint
	hopSize := windowSize / 2
	for i := 0; i < len(samples)-windowSize; i += hopSize {
		chunk := samples[i : i+windowSize]
		fftRes := signal.FFT(signal.PadDataToPowerOfTwo(signal.ApplyHanningWindow(chunk)))
		mags := signal.ComputeMagnitudes(fftRes)
		t := float64(i) / float64(data.SampleRate)
		peaks := signal.GetFingerprintPoints(mags, int(data.SampleRate), windowSize, t)
		queryPoints = append(queryPoints, peaks...)
	}

	totalPoints := len(queryPoints)
	if totalPoints == 0 {
		return MatchResult{QueryFile: filepath.Base(path), TotalPoints: 0}, nil
	}

	scores := make(map[string]map[int]int)
	for _, p := range queryPoints {
		freq := int(p.FreqHz)
		if entries, found := index[freq]; found {
			for _, entry := range entries {
				offset := entry.TimeSec - p.TimeSec
				offsetBin := int(math.Round(offset * 10))
				if scores[entry.SongName] == nil {
					scores[entry.SongName] = make(map[int]int)
				}
				scores[entry.SongName][offsetBin]++
			}
		}
	}

	bestSong := "None"
	bestOffset := 0.0
	maxScore := 0

	for song, offsetMap := range scores {
		for bin, count := range offsetMap {
			scoreWithNeighbors := count
			if v, ok := offsetMap[bin-1]; ok {
				scoreWithNeighbors += v
			}
			if v, ok := offsetMap[bin+1]; ok {
				scoreWithNeighbors += v
			}

			if scoreWithNeighbors > maxScore {
				maxScore = scoreWithNeighbors
				bestSong = song
				bestOffset = float64(bin) / 10.0
			}
		}
	}

	confidence := 0.0
	if totalPoints > 0 {
		confidence = (float64(maxScore) / float64(totalPoints))
	}

	return MatchResult{
		QueryFile:   filepath.Base(path),
		BestMatch:   filepath.Base(bestSong),
		Offset:      bestOffset,
		Score:       maxScore,
		TotalPoints: totalPoints,
		Confidence:  confidence,
		ProcessTime: time.Since(startTime),
	}, nil
}

func runSingleMode(file string, index map[int][]IndexEntry) {
	fmt.Printf("Analyzing: %s\n", file)
	res, err := identifyAudio(file, index)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nResults:")
	fmt.Printf("   Match:      %s\n", filepath.Base(res.BestMatch))
	fmt.Printf("   Offset:     %.1fs\n", res.Offset)
	fmt.Printf("   Score:      %d / %d points\n", res.Score, res.TotalPoints)
	fmt.Printf("   Confianza:  %.2f%%\n", res.Confidence)

	fmt.Println("\nBeredict:")

	if res.Confidence > ConfidenceThreshold && res.Score > 5 {
		fmt.Println("    Match found")
	} else {
		fmt.Println("    Unnable to find a match (Low confidence)")
	}
}

func runBatchMode(folder string, index map[int][]IndexEntry, csvPath string) {
	files, _ := filepath.Glob(filepath.Join(folder, "*.wav"))
	fmt.Printf("Processing %d files in '%s'\n", len(files), folder)

	csvFile, err := os.Create(csvPath)
	if err != nil {
		log.Fatal("Error generating CSV:", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Escribir cabecera
	writer.Write([]string{"Query File", "Best Match", "Offset (s)", "Score", "Total Points", "Confidence %", "Time", "Status"})

	// Procesar
	for i, file := range files {
		fmt.Printf("[%d/%d] Processing %s ... ", i+1, len(files), filepath.Base(file))

		res, err := identifyAudio(file, index)
		if err != nil {
			fmt.Println("Error")
			continue
		}

		status := "NO MATCH"
		if res.Confidence > confidenceThreshold && res.Score > 5 {
			status = "MATCH"
		}

		writer.Write([]string{
			res.QueryFile,
			res.BestMatch,
			fmt.Sprintf("%.2f", res.Offset),
			strconv.Itoa(res.Score),
			strconv.Itoa(res.TotalPoints),
			fmt.Sprintf("%.2f", res.Confidence),
			res.ProcessTime.String(),
			status,
		})
		fmt.Printf("Candidate: %s (Conf: %.2f%%)\n", res.BestMatch, res.Confidence)
	}

	fmt.Printf("\nReport saved to: %s\n", csvPath)
}

func _runBatchMode(folder string, index map[int][]IndexEntry, csvPath string) {
	files, _ := filepath.Glob(filepath.Join(folder, "*.wav"))
	totalFiles := len(files)
	fmt.Printf("Batch mode: processing %d files in '%s'\n", len(files), folder)

	jobs := make(chan string, totalFiles)
	results := make(chan MatchResult, totalFiles)

	numWorkers := runtime.NumCPU()

	var wg sync.WaitGroup

	for range numWorkers {
		wg.Add(1)
		go worker(jobs, results, index, &wg)
	}

	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	csvFile, err := os.Create(csvPath)
	if err != nil {
		log.Fatal("Error creating CSV:", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	writer.Comma = ';'
	defer writer.Flush()

	writer.Write([]string{"Query File", "Best Match", "Offset (s)", "Score", "Total Points", "Confidence %", "Time", "Status"})

	count := 0
	startTime := time.Now()

	for res := range results {
		count++

		status := "NO MATCH"
		if res.Confidence > ConfidenceThreshold && res.Score > 5 {
			status = "MATCH"
		} else if res.TotalPoints == 0 {
			status = "ERROR"
		}

		writer.Write([]string{
			res.QueryFile,
			res.BestMatch,
			fmt.Sprintf("%.2f", res.Offset),
			strconv.Itoa(res.Score),
			strconv.Itoa(res.TotalPoints),
			fmt.Sprintf("%.2f", res.Confidence),
			res.ProcessTime.String(),
			status,
		})

		printProgress(count, totalFiles, res.QueryFile, status, res.BestMatch)
	}

	fmt.Printf("\n\nProcessing finished in %v\n", time.Since(startTime))
	fmt.Printf("Report saved to: %s\n", csvPath)
}

func worker(jobs <-chan string, results chan<- MatchResult, index map[int][]IndexEntry, wg *sync.WaitGroup) {
	defer wg.Done()

	for path := range jobs {
		res, err := identifyAudio(path, index)

		if err != nil {
			results <- MatchResult{
				QueryFile: filepath.Base(path),
				Score:     0,
				BestMatch: "ERROR_READ",
			}
		} else {
			res.QueryFile = filepath.Base(res.QueryFile)
			results <- res
		}
	}
}

func printProgress(current, total int, filename, status, candidate string) {
	icon := "✅"
	if status == "NO MATCH" {
		icon = "❓"
	}
	if status == "ERROR" {
		icon = "❌"
	}

	if current < 10 {
		fmt.Printf("\n[ %d/%d] %s %s -> %s (%s)             ", current, total, icon, filename, status, candidate)
	} else {
		fmt.Printf("\n[%d/%d] %s %s -> %s (%s)             ", current, total, icon, filename, status, candidate)
	}
}
