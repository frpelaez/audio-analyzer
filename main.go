package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	sigproc "audioanalyzer/sig-proc"

	"github.com/go-audio/wav"
)

type AudioData struct {
	sampleRate int
	channels   [][]float64
}

type Config struct {
	InputFile  string
	OutputFile string
	Format     string
	WindowSize int
}

func main() {
	var cfg Config

	flag.StringVar(&cfg.OutputFile, "o", "output.csv", "Name of the output file")
	flag.StringVar(&cfg.Format, "f", "csv", "Export format (csv/bin)")
	flag.IntVar(&cfg.WindowSize, "n", 4096, "Window size for the FFT")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <audio_file.wav>\n", os.Args[0])
		fmt.Println("\nAvailable options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("ERROR: audio file not provided")
		flag.Usage()
		os.Exit(1)
	}
	cfg.InputFile = args[0]

	if cfg.Format != "csv" && cfg.Format != "bin" {
		log.Fatalf("Invalid format: %s. Use only 'cvs' or 'bin'", cfg.Format)
	}

	data, err := readWavToFloats(cfg.InputFile)
	if err != nil {
		log.Fatal(err)
	}

	windowSize := cfg.WindowSize
	hopSize := windowSize / 2
	sampleRate := data.sampleRate
	samples := data.channels[0]

	fmt.Printf("File '%s' read successfully\n", cfg.InputFile)
	fmt.Printf("Sample frequency: %d Hz \n", sampleRate)
	fmt.Printf("Channels: %d\n", len(data.channels))
	fmt.Printf("Samples per channel: %d\n", len(samples))
	fmt.Printf("Outputing results to: %s.%s\n", cfg.OutputFile, cfg.Format)
	fmt.Printf("Window size for FFT: %d\n", windowSize)

	outFile, err := os.Create(fmt.Sprintf("%s.%s", cfg.OutputFile, cfg.Format))
	if err != nil {
		log.Fatalf("Unnable to create the file: %v", err)
	}
	defer outFile.Close()

	var csvWriter *csv.Writer
	if cfg.Format == "csv" {
		csvWriter = csv.NewWriter(outFile)
		defer csvWriter.Flush()

		header := []string{"Time_Sec"}
		numBins := windowSize / 2
		for k := range numBins {
			freq := float64(k) * float64(sampleRate) / float64(windowSize)
			header = append(header, fmt.Sprintf("%.0fHz", freq))
		}
		csvWriter.Write(header)
	}

	start := time.Now()

	for i := 0; i < len(samples)-windowSize; i += hopSize {
		chunk := samples[i : i+hopSize]
		windowed := sigproc.ApplyHanningWindow(chunk)
		padded := sigproc.PadDataToPowerOfTwo(windowed)
		fftResult := sigproc.FFT(padded)
		magnitudes := sigproc.ComputeMagnitudes(fftResult)

		if cfg.Format == "csv" {
			row := make([]string, len(magnitudes)+1)
			time := float64(i) / float64(sampleRate)
			row[0] = fmt.Sprintf("%.3f", time)

			for k, val := range magnitudes {
				row[k+1] = fmt.Sprintf("%.2f", val)
			}

			csvWriter.Write(row)
		}

		// if (i/hopSize)%100 == 0 {
		// 	percent := float64(i) / float64(len(samples)) * 100
		// 	fmt.Printf("Progress: %.1f%%\n", percent)
		// }
	}

	fmt.Printf("Finished in %v.\n", time.Since(start))

	// slowFactor := 1.0
	// frameDuration := time.Duration((float64(hopSize) / float64(sampleRate)) * float64(time.Second) * slowFactor)

	// keyboardChan := make(chan bool)
	// go func() {
	// 	scanner := bufio.NewScanner(os.Stdin)
	// 	fmt.Println("Press [INTRO] to pause/continue")
	//
	// 	for scanner.Scan() {
	// 		keyboardChan <- true
	// 	}
	// }()
	//
	// fmt.Println("Initializing frequency visualizer... (Ctrl+C to quit)")
	// time.Sleep(time.Second * 2)
	//
	// isPaused := false
	//
	// i := 0
	// for i < len(samples)-windowSize {
	// 	select {
	// 	case <-keyboardChan:
	// 		isPaused = !isPaused
	// 		if isPaused {
	// 			fmt.Println("--- Paused (Press [INTRO] to continue) ---")
	// 		}
	// 	default:
	// 	}
	//
	// 	if isPaused {
	// 		time.Sleep(100 * time.Millisecond)
	// 		continue
	// 	}
	//
	// 	start := time.Now()
	//
	// 	chunk := samples[i : i+windowSize]
	// 	windowedChunk := sigproc.ApplyHanningWindow(chunk)
	// 	complexInput := sigproc.PadDataToPowerOfTwo(windowedChunk)
	// 	fftResult := sigproc.FFT(complexInput)
	// 	magnitudes := sigproc.ComputeMagnitudes(fftResult)
	//
	// 	fmt.Print("\033[H\033[2J")
	//
	// 	fmt.Printf("Time: %.1fs / %.1fs\n", float64(i)/float64(sampleRate), float64(len(samples))/float64(sampleRate))
	//
	// 	drawLogSpectrum(magnitudes, sampleRate, 15)
	//
	// 	i += hopSize
	//
	// 	elapsed := time.Since(start)
	// 	if elapsed < frameDuration {
	// 		time.Sleep(frameDuration - elapsed)
	// 	}
	// }
}

func readWavToFloats(path string) (*AudioData, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	decoder := wav.NewDecoder(f)
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("invalid wav file")
	}

	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, err
	}

	numChannels := buf.Format.NumChannels
	numSamples := len(buf.Data) / numChannels

	channels := make([][]float64, numChannels)
	for i := range channels {
		channels[i] = make([]float64, numSamples)
	}

	bitDepth := buf.SourceBitDepth
	factor := math.Pow(2, float64(bitDepth)-1)

	for i, sample := range buf.Data {
		channelIdx := i % numChannels
		sampleIdx := i / numChannels
		channels[channelIdx][sampleIdx] = float64(sample) / factor
	}

	return &AudioData{
		sampleRate: buf.Format.SampleRate,
		channels:   channels,
	}, nil
}

func drawSpectrum(dbValues []float64, numBars int) {
	chunkSize := len(dbValues) / numBars

	fmt.Println("\n--- Frequency Spectrum ---")

	for i := range numBars {
		sum := 0.0
		start := i * chunkSize
		end := min(start+chunkSize, len(dbValues))
		for j := start; j < end; j++ {
			sum += dbValues[j]
		}
		avgDB := sum / float64(end-start)

		minDB := -60.0
		maxDB := 0.0

		normalizedHeight := (avgDB - minDB) / (maxDB - minDB)
		if normalizedHeight < 0 {
			normalizedHeight = 0
		}
		if normalizedHeight > 1 {
			normalizedHeight = 1
		}

		barLen := int(normalizedHeight * 40)

		bar := strings.Builder{}
		for range barLen {
			bar.WriteString("█")
		}

		fmt.Printf("%2d | %s (%.1f dB)\n", i+1, bar.String(), avgDB)
	}
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
		fmt.Printf("%8s | %s (%.1f dB)\n", freqLabel, barBuilder.String(), dbVal)
	}
}
