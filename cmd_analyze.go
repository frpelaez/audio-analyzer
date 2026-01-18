package main

import (
	sigproc "audioanalyzer/sig-proc"
	"encoding/csv"
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
