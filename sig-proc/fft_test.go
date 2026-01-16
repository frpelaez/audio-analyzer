package sigproc

import (
	"math/rand"
	"testing"
)

func generateRandomData(n int) []complex128 {
	data := make([]complex128, n)
	for i := range n {
		data[i] = complex(rand.Float64(), rand.Float64())
	}
	return data
}

func BenchmarkFFT_4096(b *testing.B) {
	input := generateRandomData(4096)

	for b.Loop() {
		FFT(input)
	}
}

func BenchmarkFFT_1048576(b *testing.B) {
	input := generateRandomData(1048576)

	for b.Loop() {
		FFT(input)
	}
}

func BenchmarkFFTSequential_1048576(b *testing.B) {
	input := generateRandomData(1048576)

	for b.Loop() {
		fftSequential(input)
	}
}
