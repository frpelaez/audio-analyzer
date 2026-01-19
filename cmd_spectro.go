package main

import (
	sigproc "audateci/sig-proc"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func runSpectroCmd(args []string) {
	cmd := flag.NewFlagSet("spectro", flag.ExitOnError)

	outputImg := cmd.String("o", "spectrogram.png", "Name of the output image")
	pyScript := cmd.String("script", "./spectrogram/spectro.py", "Path to the python script for visualization")
	windowSize := cmd.Int("winsize", 4096, "Size of the window used for FFT (must be a power of two)")

	cmd.Parse(args)

	if cmd.NArg() < 1 {
		fmt.Println("Error. Missing audio file")
		fmt.Println("Usage: audateci spectro [options] <audio_file.wav>")
		fmt.Println("Available options are:")
		cmd.PrintDefaults()
		os.Exit(1)
	}
	audioPath := cmd.Arg(0)

	tempCSV, err := os.CreateTemp("", "spectro_data_*.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tempCSV.Name())

	fmt.Printf("Processing audio from: %s\n", audioPath)
	fmt.Printf("Generating intermediate files in: %s\n", tempCSV.Name())

	sigproc.GenerateCSV(audioPath, tempCSV, *windowSize)

	tempCSV.Close()

	fmt.Printf("Executing python script for visualization: %s\n", *pyScript)

	pythonCmd := exec.Command("python3", *pyScript, tempCSV.Name(), "--save", *outputImg)
	pythonCmd.Stdout = os.Stdout
	pythonCmd.Stderr = os.Stderr

	pyerr := pythonCmd.Run()
	if pyerr != nil {
		log.Fatalf("Error during execution of python script: %v\n", *pyScript)
	}
}
