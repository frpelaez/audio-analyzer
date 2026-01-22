package signal

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/cmplx"
	"os"

	"github.com/go-audio/wav"
)

type AudioData struct {
	SampleRate int
	Channels   [][]float64
}

func nextPowerOfTwo(n int) int {
	pow := 1
	for n > pow {
		pow *= 2
	}

	return pow
}

func MagnitudesToDB(magnitudes []float64) []float64 {
	epsilon := 1e-9

	dbValues := make([]float64, len(magnitudes))
	for i, mag := range magnitudes {
		mag = max(mag, epsilon)
		dbValues[i] = 20 * math.Log10(mag)
	}

	return dbValues
}

func PadDataToPowerOfTwo(data []float64) []complex128 {
	length := len(data)
	target := nextPowerOfTwo(length)

	padded := make([]complex128, target)
	for i, v := range data {
		padded[i] = complex(v, 0)
	}

	return padded
}

func ComputeMagnitudes(spectrum []complex128) []float64 {
	n := len(spectrum)
	magnitudes := make([]float64, n)
	for i := 0; i < n/2; i++ {
		magnitudes[i] = cmplx.Abs(spectrum[i]) * 2.0 / float64(n)
	}

	return magnitudes
}

func ApplyHanningWindow(input []float64) []float64 {
	n := len(input)
	output := make([]float64, n)
	for i := range input {
		factor := 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n-1)))
		output[i] = input[i] * factor
	}

	return output
}

func ReadWavToFloats(path string) (*AudioData, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	decoder := wav.NewDecoder(f)
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("invalid wav file")
	}

	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, err
	}

	numChannels := buf.Format.NumChannels
	numSamples := len(buf.Data) / numChannels

	channels := make([][]float64, numChannels)
	for i := range channels {
		channels[i] = make([]float64, numSamples)
	}

	bitDepth := buf.SourceBitDepth
	factor := math.Pow(2, float64(bitDepth)-1)

	for i, sample := range buf.Data {
		channelIdx := i % numChannels
		sampleIdx := i / numChannels
		channels[channelIdx][sampleIdx] = float64(sample) / factor
	}

	return &AudioData{
		SampleRate: buf.Format.SampleRate,
		Channels:   channels,
	}, nil
}

func GenerateCSV(audioPath string, file *os.File, winSize int) {
	data, err := ReadWavToFloats(audioPath)
	if err != nil {
		log.Fatal(err)
	}
	samples := data.Channels[0]
	sampleRate := data.SampleRate

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Time_Sec"}
	for k := range winSize / 2 {
		freq := float64(k) * float64(sampleRate) / float64(winSize)
		header = append(header, fmt.Sprintf("%.0fHz", freq))
	}
	writer.Write(header)

	hopSize := winSize / 2
	for i := 0; i < len(samples)-winSize; i += hopSize {
		chunk := samples[i : i+hopSize]
		windowed := ApplyHanningWindow(chunk)
		padded := PadDataToPowerOfTwo(windowed)
		fftRes := FFT(padded)
		mags := ComputeMagnitudes(fftRes)

		row := make([]string, len(mags)+1)
		row[0] = fmt.Sprintf("%.3f", float64(i)/float64(sampleRate))
		for k, v := range mags {
			row[k+1] = fmt.Sprintf("%.2f", v)
		}
		writer.Write(row)
	}
}
