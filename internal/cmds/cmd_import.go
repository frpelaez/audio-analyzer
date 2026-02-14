package cmd

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	signal "audateci/internal/signal"
)

func RunImportCmd(args []string) {
	cmd := flag.NewFlagSet("import", flag.ExitOnError)
	outputDir := cmd.String("o", "db", "Output directory for the fingerprints")
	cmd.Parse(args)

	if cmd.NArg() < 1 {
		fmt.Println("Usage: audateci import -o <output-dir> <playlist-info-file.txt>")
		os.Exit(1)
	}

	inputFile := cmd.Arg(0)

	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Unnable to open song-list-file: %v", err)
	}
	defer file.Close()

	var songs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "//") {
			songs = append(songs, line)
		}
	}

	if len(songs) == 0 {
		fmt.Println("File is empty.")
		return
	}

	for i, query := range songs {
		safeFilename := sanitizeFilename(query)
		jsonPath := filepath.Join(*outputDir, safeFilename+".json")

		if _, err := os.Stat(jsonPath); err == nil {
			fmt.Printf("[%d/%d] Skipping (already exists): %s\n", i+1, len(songs), query)
			continue
		}

		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(songs), query)

		tempWav := fmt.Sprintf("temp_%d.wav", i)

		dlCmd := exec.Command("yt-dlp",
			"ytsearch1:"+query,
			"-x",
			"--audio-format", "wav",
			"-o", tempWav,
			"--force-overwrites",
		)

		if err := dlCmd.Run(); err != nil {
			fmt.Printf("   Error dowloading: %v\n", err)
			continue
		}

		fingerprint := processAudioToFingerprint(tempWav, query)

		if len(fingerprint.Points) == 0 {
			fmt.Printf("   Warninga: no audio data found in %s\n", tempWav)
			os.Remove(tempWav)
			continue
		}

		jsonFile, _ := os.Create(jsonPath)
		json.NewEncoder(jsonFile).Encode(fingerprint)
		jsonFile.Close()

		fmt.Printf("   Fingerprint saved (%d points)\n", len(fingerprint.Points))

		os.Remove(tempWav)

		time.Sleep(2 * time.Second)
	}

	fmt.Println("\nImport completed")
}

func sanitizeFilename(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		name = strings.ReplaceAll(name, char, "")
	}
	return name
}

func processAudioToFingerprint(wavPath, originalName string) FingerprintFile {
	data, err := signal.ReadWavToFloats(wavPath)
	if err != nil {
		log.Println("Error leyendo wav:", err)
		return FingerprintFile{}
	}
	samples := data.Channels[0]

	windowSize := 2048
	hopSize := windowSize / 2
	var points []signal.KeyPoint

	for i := 0; i < len(samples)-windowSize; i += hopSize {
		chunk := samples[i : i+windowSize]
		fftRes := signal.FFT(signal.PadDataToPowerOfTwo(signal.ApplyHanningWindow(chunk)))
		mags := signal.ComputeMagnitudes(fftRes)
		t := float64(i) / float64(data.SampleRate)
		peaks := signal.GetFingerprintPoints(mags, int(data.SampleRate), windowSize, t)
		points = append(points, peaks...)
	}

	return FingerprintFile{
		Filename: originalName,
		Points:   points,
	}
}

func RunFingerprintDir(args []string) {
	cmd := flag.NewFlagSet("fpdir", flag.ExitOnError)
	outputDir := cmd.String("o", "fdb", "Output directory for the fingerprints")
	cmd.Parse(args)

	if cmd.NArg() < 1 {
		fmt.Println("Usage: audateci fpdir -o <output-dir> <dir-with-wavs>")
		os.Exit(1)
	}

	inputFolder := cmd.Arg(0)

	files, _ := filepath.Glob(filepath.Join(inputFolder, "*.wav"))
	totalFiles := len(files)
	fmt.Printf("Processing %d files...\n", totalFiles)
	for i, file := range files {
		name := strings.SplitN(filepath.Base(file), ".", 2)[0]
		fmt.Printf("[%d/%d] Processing '%s'...\n", i+1, totalFiles, name)
		fpFile := processAudioToFingerprint(file, name)

		if len(fpFile.Points) == 0 {
			fmt.Printf("Warning: no audio data found in %s\n", file)
			continue
		}

		outputFile := filepath.Join(*outputDir, sanitizeFilename(name)+".json")
		jsonFile, _ := os.Create(outputFile)
		json.NewEncoder(jsonFile).Encode(fpFile)
		jsonFile.Close()

		fmt.Printf("Fingerprint saved to '%s' (%d points)\n", outputFile, len(fpFile.Points))
	}

	fmt.Println("Process finished successfully")
}
