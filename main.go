package main

import (
	cmds "audateci/internal/cmds"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "listen":
		cmds.RunListenCmd(args)
	case "analyze":
		cmds.RunAnalyzeCmd(args)
	case "spectro":
		cmds.RunSpectroCmd(args)
	case "fingerprint":
		cmds.RunFingerprintCmd(args)
	case "match":
		cmds.RunMatchCmd(args)
	case "identify":
		cmds.RunIdentifyCmd(args)
	case "-h", "--help", "help":
		printHelp()
	default:
		fmt.Printf("Unkown command: '%s'\n", command)
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
	fmt.Println("  match         Decide if two fingerprints have a match and what is the offset between them")
	fmt.Println("  identify      Run a match between a given audio file and a directory containing audio fingerprints")
	fmt.Println("\nType audateci <command> -h for specific help")
}
