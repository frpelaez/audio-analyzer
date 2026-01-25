package main

import (
	cmd "audateci/cmd"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "listen":
		cmd.RunListenCmd(os.Args[2:])
	case "analyze":
		cmd.RunAnalyzeCmd(os.Args[2:])
	case "spectro":
		cmd.RunSpectroCmd(os.Args[2:])
	case "fingerprint":
		cmd.RunFingerprintCmd(os.Args[2:])
	case "-h", "--help", "help":
		printHelp()
	default:
		fmt.Printf("Unkown command: '%s'\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Usage: audateci <command> [options] <audio_file.wav>")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  analyze       Analyze the audio file and export data to csv")
	fmt.Println("  fingerprint   Calculate the audio fingerprint of wav file and export it to json format")
	fmt.Println("  listen        Visualize the frequencies contained in the audio file")
	fmt.Println("  spectro       Compute spectrogram from audio file and export to png")
	fmt.Println("\nType audateci <command> -h for specific help")
}
