package main

import (
	cmds "audateci/internal/cmds"
	"fmt"
	"os"

	"github.com/fatih/color"
)

func main() {
	if len(os.Args) < 2 {
		cmds.RunReplCmd()
		return
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
	case "repl":
		cmds.RunReplCmd()
	case "-h", "--help", "help":
		printHelp()
	default:
		fmt.Printf("Unkown command: '%s'\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	usageStyle := color.New(color.FgMagenta, color.Bold)
	audateci := color.BlueString(" audateci ")
	command := color.CyanString("<command> ")
	options := color.GreenString("[options] ")
	file := color.CyanString("<audio-file.wav>")
	usageStyle.Print("Usage:")
	println(audateci + command + options + file)

	println("\n    Simple tool for analyzing audio files via their time-frequency decomposition      ")

	avCmdStyle := color.New(color.FgMagenta, color.Bold)
	avCmdStyle.Println("\nAvailable commands:")

	cmdsStyle := color.New(color.FgCyan)
	println(cmdsStyle.Sprint("    analyze") + "        Analyze the audio file and export data to csv")
	println(cmdsStyle.Sprint("    fingerprint") + "    Calculate the audio fingerprint of wav file and export it to json format")
	println(cmdsStyle.Sprint("    identify") + "       Run a match between a given audio file and a directory containing audio fingerprints")
	println(cmdsStyle.Sprint("    listen") + "         Visualize the frequencies contained in the audio file")
	println(cmdsStyle.Sprint("    match") + "          Decide if two fingerprints have a match and what is the offset between them")
	println(cmdsStyle.Sprint("    repl") + "           Run the audateci repl")
	println(cmdsStyle.Sprint("    spectro") + "        Compute spectrogram from audio file and export to png")

	println("\nType " + color.BlueString("audateci ") + color.CyanString("<command> ") + color.GreenString("-h") + " for specific help\n")
}
