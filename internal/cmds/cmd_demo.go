package cmd

import (
	"fmt"
	"log"
	"path/filepath"
)

func RunDemoCmd() {
	dbPath := "./db"
	firstSongPath := "./fragments_test/nat-king-cole-love.wav"
	secondSongPath := "./fragments_test/misterio-nada-sospechoso.wav"

	fmt.Printf("Cargando base de datos de canciones %s...\n", dbPath)
	dbIndex := loadDatabase(dbPath)

	fmt.Printf("Identificando primera canción: %s\n", filepath.Base(firstSongPath))
	firstRes, err := identifyAudio(firstSongPath, dbIndex)
	if err != nil {
		log.Fatal(err)
	}

	songTitle := filepath.Base(firstRes.BestMatch)
	url, err := getVideoURL(songTitle)
	if err != nil {
		fmt.Printf("Unnable to get url for best matching song: '%s'", songTitle)
		url = "NONE"
	}

	fmt.Printf("   Resultado:       %s (%s)\n", songTitle, url)
	fmt.Printf("   Desfase:         %.1fs\n", firstRes.Offset)
	fmt.Printf("   Puntuación (%%):  %.2f%%\n", firstRes.Confidence)

	err = openURL(url)
	if err != nil {
		fmt.Printf("Unnable to open song url: %v", err)
	}

	fmt.Printf("\nIdentificando segunda canción: ???\n")
	secondRes, err := identifyAudio(secondSongPath, dbIndex)
	if err != nil {
		log.Fatal(err)
	}

	songTitle = filepath.Base(secondRes.BestMatch)
	url, err = getVideoURL(songTitle)
	if err != nil {
		fmt.Printf("Unnable to get url for best matching song: '%s'", songTitle)
		url = "NONE"
	}

	url += fmt.Sprintf("&autoplay=1&start=%d", int(secondRes.Offset))

	fmt.Printf("   Resultado:       %s (%s)\n", songTitle, url)
	fmt.Printf("   Desfase:         %.1fs\n", secondRes.Offset)
	fmt.Printf("   Puntuación (%%):  %.2f%%\n", secondRes.Confidence)

	err = openURL(url)
	if err != nil {
		fmt.Printf("Unnable to open song url: %v", err)
	}
}
