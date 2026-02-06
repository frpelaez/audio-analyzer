package cmd

import (
	"audateci/internal/signal"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Session struct {
	TargetFile   string
	Points       []signal.KeyPoint
	SampleRate   int
	IsReady      bool
	AnalysisTime time.Duration
	mu           sync.Mutex
}

func RunReplCmd() {
	reader := bufio.NewReader(os.Stdin)

	var audioPath string

	for {
		fmt.Print("Introduce the path to the audio file you wnat to analyze: ")
		input, _ := reader.ReadString('\n')
		audioPath = strings.TrimSpace(input)

		if _, err := os.Stat(audioPath); err == nil {
			break
		}

		fmt.Println("❌ File does not exist or it cannot be accessed. Try again")
	}

	session := &Session{
		TargetFile: audioPath,
		IsReady:    false,
	}

	go processAudioBackground(session)

	fmt.Println("✅ Audio loaded. Processing in the background...")
	fmt.Println("You can start typing commands now")

	for {
		fmt.Print(">>> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		switch command {
		case "identify":
			handleIdentify(session, args)
		case "status":
			handleStatus(session)
		case "exit", "quit":
			fmt.Print("Exiting session")
			return
		case "help":
			printReplHelp()
		default:
			fmt.Printf("❌ Unkown command '%s'. Type 'help' to see the available options", command)
		}
	}
}

func processAudioBackground(s *Session) {
	startTime := time.Now()

	data, err := signal.ReadWavToFloats(s.TargetFile)
	if err != nil {
		log.Printf("\n ❌ Error reading audio file in the background: %v\n>>> ", err)
	}
	samples := data.Channels[0]

	windowSize := 2048
	hopSize := windowSize / 2
	var foundPoints []signal.KeyPoint

	for i := 0; i < len(samples)-windowSize; i += hopSize {
		chunk := samples[i : i+windowSize]
		windowed := signal.ApplyHanningWindow(chunk)
		padded := signal.PadDataToPowerOfTwo(windowed)
		fftRes := signal.FFT(padded)
		mags := signal.ComputeMagnitudes(fftRes)

		time := float64(i) / float64(data.SampleRate)
		peaks := signal.GetFingerprintPoints(mags, data.SampleRate, windowSize, time)
		foundPoints = append(foundPoints, peaks...)
	}

	s.mu.Lock()
	s.Points = foundPoints
	s.SampleRate = data.SampleRate
	s.IsReady = true
	s.AnalysisTime = time.Since(startTime)
	s.mu.Unlock()

	fmt.Printf("\n✅ Analysis completed: found %d keypoints in %v\n>>> ", len(foundPoints), time.Since(startTime))
}

func handleStatus(s *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IsReady {
		fmt.Printf("Audio file:        %s\nStatus:            processed (%d keypoints)\nAnalysis duration: %v\n",
			s.TargetFile, len(s.Points), s.AnalysisTime)
	} else {
		fmt.Printf("Audio file: %s\nStatus:     Processing... please wait\n", s.TargetFile)
	}
}

func handleIdentify(s *Session, args []string) {
	s.mu.Lock()
	if !s.IsReady {
		fmt.Println("❌ Audio analysis has not finished yet. Wait a moment")
		s.mu.Unlock()
		return
	}

	points := s.Points
	s.mu.Unlock()

	if len(args) < 1 {
		fmt.Println("❌ Usage: identify <dir-with-fingerprints>")
		return
	}
	dbFolder := args[0]

	fmt.Println("Looking for matches in the data base...")

	dbIndex := make(map[int][]IndexEntry)
	files, err := filepath.Glob(filepath.Join(dbFolder, "*.json"))
	if err != nil || len(files) == 0 {
		fmt.Println("❌ No fingerprints (.json) found in the directory")
		return
	}

	for _, file := range files {
		f, _ := os.Open(file)
		var fingerprint FingerprintFile
		json.NewDecoder(f).Decode(&fingerprint)
		f.Close()
		for _, p := range fingerprint.Points {
			freq := int(p.FreqHz)
			dbIndex[freq] = append(dbIndex[freq], IndexEntry{SongName: fingerprint.Filename, TimeSec: p.TimeSec})
		}
	}

	scores := make(map[string]map[int]int)
	for _, pFrag := range points {
		freq := int(pFrag.FreqHz)
		if entries, found := dbIndex[freq]; found {
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
			if v, ok := offsetMap[bin-1]; ok {
				neighborsScore += v
			}
			if v, ok := offsetMap[bin-1]; ok {
				neighborsScore += v
			}

			if neighborsScore > bestScore {
				bestSong = song
				bestOffset = float64(bin) / 10.0
				bestScore = neighborsScore
			}
		}
	}

	fmt.Println("Results:")
	if bestScore > 100 {
		fmt.Println("   ✅ Match found!")
		fmt.Printf("   Song:   %s\n", bestSong)
		fmt.Printf("   Offset: %.1fs\n", bestOffset)
		fmt.Printf("   Score:  %d matches\n", bestScore)
	} else {
		fmt.Printf("   ❌ No clear matches with decision threshold %d\n", 100)
	}
}

func printReplHelp() {
	fmt.Println("Available commands:")
	fmt.Println("    identify <directory>  Compare loaded audio with all the fingerprints contained in <dir>")
	fmt.Println("    status                Check the status of the analysis running in the background")
	fmt.Println("    exit                  Exit the program")
}
