package bench

import (
	"fmt"
	"time"

	"golang.org/x/exp/rand"

	"gonum.org/v1/gonum/stat/sampleuv"
)

type InvocationSampler func([]Invocations) []float64

func AllInvocations(ivs []Invocations) []float64 {
	if len(ivs) == 0 {
		return make([]float64, 0)
	}

	var out []float64
	for _, iv := range ivs {
		for i := 0; i < iv.Count; i++ {
			out = append(out, iv.Value)
		}
	}
	return out
}

func MeanInvocations(ivs []Invocations) []float64 {
	var total int
	var sum float64

	for _, iv := range ivs {
		total += iv.Count
		sum += float64(iv.Count) * iv.Value
	}

	mean := sum / float64(total)
	return []float64{mean}
}

func SampleInvocations(samples int) InvocationSampler {
	return func(ivs []Invocations) []float64 {
		livs := len(ivs)
		weights := make([]float64, livs)
		values := make([]float64, livs)
		var totalValues int
		for i := 0; i < livs; i++ {
			iv := ivs[i]
			totalValues += iv.Count
			weights[i] = float64(iv.Count)
			values[i] = iv.Value
		}

		// if samples is equal or greater than the invocations passed
		if totalValues <= samples {
			return AllInvocations(ivs)
		}

		sampler := sampleuv.NewWeighted(weights, rand.NewSource(uint64(time.Now().UnixNano())))

		out := make([]float64, samples)
		for i := 0; i < samples; i++ {
			idx, ok := sampler.Take()
			if !ok {
				panic(fmt.Sprintf("Could not take sample nr %d. Should never happen", i))
			}
			out[i] = values[idx]
			sampler.Reweight(idx, weights[idx])
		}
		return out
	}
}
