package main

import (
	sigproc "audioanalyzer/sig-proc"
	// "encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func runAnalyzeCmd(args []string) {
	cmd := flag.NewFlagSet("analyze", flag.ExitOnError)

	output := cmd.String("o", "output.csv", "Output file")
	format := cmd.String("format", "csv", "Output format (csv/bin (wip))")
	winsize := cmd.Int("winsize", 4096, "Window size used for the FFT (must be a power of two)")

	cmd.Parse(args)

	if cmd.NArg() < 1 {
		fmt.Println("Error. Missing audio file")
		fmt.Println("Usage: audateci analyze [options] <audio_file.wav>")
		fmt.Println("Available options are:")
		cmd.PrintDefaults()
		os.Exit(1)
	}
	inputFile := cmd.Arg(0)

	data, err := sigproc.ReadWavToFloats(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	samples := data.Channels[0]

	fmt.Printf("File '%s' read successfully\n", inputFile)
	fmt.Printf("Sample frequency: %d Hz \n", data.SampleRate)
	fmt.Printf("Channels: %d\n", len(data.Channels))
	fmt.Printf("Samples per channel: %d\n", len(samples))
	fmt.Printf("Window size for FFT: %d\n", *winsize)
	fmt.Printf("Outputing results to: %s.%s\n", *output, *format)

	outFile, err := os.Create(fmt.Sprintf("%s.%s", *output, *format))
	if err != nil {
		log.Fatalf("Unnable to create the file: %v", err)
	}
	defer outFile.Close()

	start := time.Now()
	sigproc.GenerateCSV(inputFile, outFile, *winsize)
	fmt.Printf("Finished in %.3fs\n", time.Since(start).Seconds())
}
