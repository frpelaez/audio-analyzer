package cmd

import (
	"audateci/signal"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

type AudioFingerprint struct {
	Filename   string            `json:"filename"`
	Duration   float64           `json:"duration"`
	SampleRate int               `json:"sample_rate"`
	Points     []signal.KeyPoint `json:"points"`
}

func RunFingerprintCmd(args []string) {
	cmd := flag.NewFlagSet("fingerprint", flag.ExitOnError)
	output := cmd.String("o", "fingerprint.json", "Output file (.json)")
	windowSize := cmd.Int("winsize", 2048, "Window size used for fft, must be a power of two")

	cmd.Parse(args)

	if cmd.NArg() < 1 {
		fmt.Println("Error. Missing audio file")
		fmt.Println("Usage: audateci listen [options] <audio_file.wav>")
		fmt.Println("Available options are:")
		cmd.PrintDefaults()
		os.Exit(1)
	}
	inputFile := cmd.Arg(0)

	fmt.Printf("Generating fingerprint for '%s'\n", inputFile)
	data, err := signal.ReadWavToFloats(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	samples := data.Channels[0]

	var resultPoints []signal.KeyPoint
	hopSize := *windowSize / 2

	for i := 0; i < len(samples)-*windowSize; i += hopSize {
		chunk := samples[i : i+*windowSize]
		windowed := signal.ApplyHanningWindow(chunk)
		padded := signal.PadDataToPowerOfTwo(windowed)
		fftRes := signal.FFT(padded)
		magnitudes := signal.ComputeMagnitudes(fftRes)

		currentTime := float64(i) / float64(data.SampleRate)

		peaks := signal.GetFingerprintPoints(magnitudes, data.SampleRate, *windowSize, currentTime)

		resultPoints = append(resultPoints, peaks...)
	}

	fingerprintData := AudioFingerprint{
		Filename:   inputFile,
		Duration:   float64(len(samples)) / float64(data.SampleRate),
		SampleRate: data.SampleRate,
		Points:     resultPoints,
	}

	file, _ := os.Create(*output)
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	jsonerr := encoder.Encode(fingerprintData)
	if jsonerr != nil {
		log.Fatal("Error saving JSON:", err)
	}

	fmt.Printf("Audio fingerprint saved successfully to '%s'. Found %d key points", *output, len(resultPoints))
}
