package sigproc

import (
	"math"
	"math/cmplx"
)

func nextPowerOfTwo(n int) int {
	if n == 0 {
		return 1
	}

	pow := 2
	for n > pow {
		pow *= 2
	}

	return pow
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
