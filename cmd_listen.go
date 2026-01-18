package main

import (
	sigproc "audioanalyzer/sig-proc"
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

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
