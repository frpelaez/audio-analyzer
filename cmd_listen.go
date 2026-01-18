package main

import (
	sigproc "audioanalyzer/sig-proc"
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
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

	f, err := os.Open(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := wav.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}

	data, err := readWavToFloats(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	windowSize := *winSize
	// hopSize := windowSize / 2
	sampleRate := data.sampleRate
	samples := data.channels[0]

	fmt.Printf("File '%s' read successfully\n", inputFile)
	fmt.Printf("Sample frequency: %d Hz \n", sampleRate)
	fmt.Printf("Channels: %d\n", len(data.channels))
	fmt.Printf("Samples per channel: %d\n", len(samples))
	fmt.Printf("Window size for FFT: %d\n", windowSize)

	// slowFactor := 1.0
	// frameDuration := time.Duration((float64(hopSize) / float64(sampleRate)) * float64(time.Second) * slowFactor)

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

	speaker.Play(ctrl)

	framesTarget := 20
	ticker := time.NewTicker(time.Second / time.Duration(framesTarget))
	defer ticker.Stop()

	startTime := time.Now()
	pausedTime := time.Duration(0)
	isPaused := false
	pausedStart := time.Now()

	for {
		select {
		case <-keyboardChan:
			isPaused = !isPaused

			speaker.Lock()
			ctrl.Paused = isPaused
			speaker.Unlock()

			if isPaused {
				pausedStart = time.Now()
				fmt.Println("--- Paused (Press [INTRO] to continue) ---")
			} else {
				pausedTime += time.Since(pausedStart)
				startTime = startTime.Add(time.Since(pausedStart))
			}

		case <-ticker.C:
			if isPaused {
				continue
			}

			elapsed := time.Since(startTime)
			sampleIdx := int(elapsed.Seconds() * float64(sampleRate))
			if sampleIdx >= len(samples)-windowSize {
				fmt.Println("\nAudio reproduction ended normally")
				return
			}

			chunk := samples[sampleIdx : sampleIdx+windowSize]
			windowedChunk := sigproc.ApplyHanningWindow(chunk)
			complexInput := sigproc.PadDataToPowerOfTwo(windowedChunk)
			fftResult := sigproc.FFT(complexInput)
			magnitudes := sigproc.ComputeMagnitudes(fftResult)

			fmt.Print("\033c\033[3J")

			drawLogSpectrum(magnitudes, sampleRate, *bars)

			percent := float64(sampleIdx) / float64(len(samples)) * 100
			fmt.Printf("\n\n %.1f%% - %.1f/%.1fs\n", percent, elapsed.Seconds(), float64(len(samples))/float64(sampleRate))
		}
	}
}
