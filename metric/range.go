package metric

import "math"

type bktRange struct {
	min float64
	max float64
}

func (rng bktRange) exponentialBuckets() []float64 {
	const maxMagnitudes = 15.0

	// Log_b(x)  = Log_a(x)/Log_a(b)
	logbase := func(x, base float64) float64 {
		return math.Log2(x) / math.Log2(base)
	}

	factor := 2.0
	magnitudes := 0.0
	for {
		max := logbase(rng.max, factor)
		min := logbase(rng.min, factor)
		magnitudes = max - min
		if magnitudes < maxMagnitudes {
			break
		}
		factor += 1.0
	}

	var buckets []float64
	for i := 0.0; i <= magnitudes+1.0; i++ {
		buckets = append(buckets, rng.min*(math.Pow(factor, i)))
	}
	return buckets
}
