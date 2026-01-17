package main

import (
	"bufio"
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
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "listen":
		runListenCmd(os.Args[2:])
	case "analyze":
		runAnalyzeCmd(os.Args[2:])
	case "-h", "--help", "help":
		printHelp()
	default:
		fmt.Printf("Unkown command: '%s'\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Usage: audateci <command> [options]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  listen    Visualize the frequencies contained in the audio file")
	fmt.Println("  analyze   Analyze the audio file and export data to CSV/Bin (wip)")
	fmt.Println("\nType audateci <command> -h for specific help")
}

func runListenCmd(args []string) {
	cmd := flag.NewFlagSet("listen", flag.ExitOnError)

	bars := cmd.Int("bars", 20, "Number of frquency bars to show")
	winSize := cmd.Int("winsize", 4096, "Window size used for FFT, must be a power of two")

	cmd.Parse(args)

	if cmd.NArg() < 1 {
		fmt.Println("Error. Missing audio file")
		fmt.Println("Usage: audateci listen [options] <audio_file.wav>")
		cmd.PrintDefaults()
		os.Exit(1)
	}
	inputFile := cmd.Arg(0)

	data, err := readWavToFloats(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	windowSize := *winSize
	hopSize := windowSize / 2
	sampleRate := data.sampleRate
	samples := data.channels[0]

	fmt.Printf("File '%s' read successfully\n", inputFile)
	fmt.Printf("Sample frequency: %d Hz \n", sampleRate)
	fmt.Printf("Channels: %d\n", len(data.channels))
	fmt.Printf("Samples per channel: %d\n", len(samples))
	fmt.Printf("Window size for FFT: %d\n", windowSize)

	slowFactor := 1.0
	frameDuration := time.Duration((float64(hopSize) / float64(sampleRate)) * float64(time.Second) * slowFactor)

	keyboardChan := make(chan bool)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Press [INTRO] to pause/continue")

		for scanner.Scan() {
			keyboardChan <- true
		}
	}()

	fmt.Println("Initializing frequency visualizer... (Ctrl+C to quit)")
	time.Sleep(time.Second * 2)

	isPaused := false

	i := 0
	for i < len(samples)-windowSize {
		select {
		case <-keyboardChan:
			isPaused = !isPaused
			if isPaused {
				fmt.Println("--- Paused (Press [INTRO] to continue) ---")
			}
		default:
		}

		if isPaused {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		start := time.Now()

		chunk := samples[i : i+windowSize]
		windowedChunk := sigproc.ApplyHanningWindow(chunk)
		complexInput := sigproc.PadDataToPowerOfTwo(windowedChunk)
		fftResult := sigproc.FFT(complexInput)
		magnitudes := sigproc.ComputeMagnitudes(fftResult)

		fmt.Print("\033[H\033[2J")

		fmt.Printf("Time: %.1fs / %.1fs\n", float64(i)/float64(sampleRate), float64(len(samples))/float64(sampleRate))

		drawLogSpectrum(magnitudes, sampleRate, *bars)

		i += hopSize

		elapsed := time.Since(start)
		if elapsed < frameDuration {
			time.Sleep(frameDuration - elapsed)
		}
	}
}

func runAnalyzeCmd(args []string) {
	cmd := flag.NewFlagSet("analyze", flag.ExitOnError)

	output := cmd.String("o", "output.csv", "Output file")
	format := cmd.String("format", "csv", "Output format (csv/bin (wip))")
	winsize := cmd.Int("winsize", 4096, "Window size used for the FFT (must be a power of two)")

	cmd.Parse(args)

	if cmd.NArg() < 1 {
		fmt.Println("Error. Missing audio file")
		fmt.Println("Usage: audateci analyze [options] <audio_file.wav>")
		cmd.PrintDefaults()
		os.Exit(1)
	}
	inputFile := cmd.Arg(0)

	data, err := readWavToFloats(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	samples := data.channels[0]
	hopSize := *winsize / 2

	fmt.Printf("File '%s' read successfully\n", inputFile)
	fmt.Printf("Sample frequency: %d Hz \n", data.sampleRate)
	fmt.Printf("Channels: %d\n", len(data.channels))
	fmt.Printf("Samples per channel: %d\n", len(samples))
	fmt.Printf("Window size for FFT: %d\n", *winsize)
	fmt.Printf("Outputing results to: %s.%s", *output, *format)

	outFile, err := os.Create(fmt.Sprintf("%s.%s", *output, *format))
	if err != nil {
		log.Fatalf("Unnable to create the file: %v", err)
	}
	defer outFile.Close()

	var csvWriter *csv.Writer
	if *format == "csv" {
		csvWriter = csv.NewWriter(outFile)
		defer csvWriter.Flush()

		header := []string{"Time_Sec"}
		numBins := *winsize / 2
		for k := range numBins {
			freq := float64(k) * float64(data.sampleRate) / float64(*winsize)
			header = append(header, fmt.Sprintf("%.0fHz", freq))
		}
		csvWriter.Write(header)
	}

	start := time.Now()

	for i := 0; i < len(samples)-*winsize; i += hopSize {
		chunk := samples[i : i+hopSize]
		windowed := sigproc.ApplyHanningWindow(chunk)
		padded := sigproc.PadDataToPowerOfTwo(windowed)
		fftResult := sigproc.FFT(padded)
		magnitudes := sigproc.ComputeMagnitudes(fftResult)

		if *format == "csv" {
			row := make([]string, len(magnitudes)+1)
			time := float64(i) / float64(data.sampleRate)
			row[0] = fmt.Sprintf("%.3f", time)

			for k, val := range magnitudes {
				row[k+1] = fmt.Sprintf("%.2f", val)
			}

			csvWriter.Write(row)
		}
	}

	fmt.Printf("Finished in %v.\n", time.Since(start))
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
