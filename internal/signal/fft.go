package signal

import (
	"math"
	"math/cmplx"
)

const ConcurrencyThreshold = 1024

func FFT(x []complex128) []complex128 {
	n := len(x)

	if n <= ConcurrencyThreshold {
		return fftSequential(x)
	}

	even := make([]complex128, n/2)
	odd := make([]complex128, n/2)
	for i := range n / 2 {
		even[i] = x[2*i]
		odd[i] = x[2*i+1]
	}

	resultEvenChannel := make(chan []complex128)
	go func() {
		resultEvenChannel <- FFT(even)
	}()

	resOdd := FFT(odd)
	resEven := <-resultEvenChannel

	result := make([]complex128, n)
	for k := range n / 2 {
		rotation := -2.0 * math.Pi * float64(k) / float64(n)
		w := cmplx.Exp(complex(0, rotation))
		t := w * resOdd[k]
		result[k] = resEven[k] + t
		result[k+n/2] = resEven[k] - t
	}

	return result
}

func fftSequential(x []complex128) []complex128 {
	n := len(x)

	if n == 1 {
		return x
	}

	even := make([]complex128, n/2)
	odd := make([]complex128, n/2)
	for i := 0; i < n/2; i++ {
		even[i] = x[2*i]
		odd[i] = x[2*i+1]
	}

	evenFft := fftSequential(even)
	oddFft := fftSequential(odd)

	result := make([]complex128, n)
	for k := 0; k < n/2; k++ {
		rotation := -2.0 * math.Pi * float64(k) / float64(n)
		w := cmplx.Exp(complex(0, rotation))
		t := w * oddFft[k]
		result[k] = evenFft[k] + t
		result[k+n/2] = evenFft[k] - t
	}

	return result
}
