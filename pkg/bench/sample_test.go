package bench_test

import (
	"math"
	"testing"

	"github.com/chrstphlbr/pa/pkg/bench"
)

func TestSampleAllInvocationsEmpty(t *testing.T) {
	ivs := []bench.Invocations{}

	aivs := bench.AllInvocations(ivs)

	if l := len(aivs); l != 0 {
		t.Fatalf("Expected no elements, but was %d", l)
	}
}

func checkAll(t *testing.T, aivs []float64) {
	if l := len(aivs); l != 35 {
		t.Fatalf("Unexpected invocations length: expected %d, was %d", 35, l)
	}

	for i, iv := range aivs {
		if i < 5 {
			if iv != 4 {
				t.Fatalf("Unexpected value at pos %d: expected %f, was %f", i, 4.0, iv)
			}
		} else if i < 15 {
			if iv != 5 {
				t.Fatalf("Unexpected value at pos %d: expected %f, was %f", i, 5.0, iv)
			}
		} else if i < 35 {
			if iv != 6 {
				t.Fatalf("Unexpected value at pos %d: expected %f, was %f", i, 6.0, iv)
			}
		}
	}
}

func TestSampleAllInvocations(t *testing.T) {
	ivs := []bench.Invocations{
		{Count: 5, Value: 4},
		{Count: 10, Value: 5},
		{Count: 20, Value: 6},
	}

	aivs := bench.AllInvocations(ivs)
	checkAll(t, aivs)
}

func TestSampleMeanEmpty(t *testing.T) {
	ivs := []bench.Invocations{}

	aivs := bench.MeanInvocations(ivs)

	if l := len(aivs); l != 1 {
		t.Fatalf("Expected 1 element, the mean, but was %d", l)
	}

	if el := aivs[0]; !math.IsNaN(el) {
		t.Fatalf("Expected NaN, because no invocations were passed, but got %f", el)
	}
}

func TestSampleMean(t *testing.T) {
	ivs := []bench.Invocations{
		{Count: 5, Value: 4},
		{Count: 10, Value: 5},
		{Count: 20, Value: 6},
	}

	aivs := bench.MeanInvocations(ivs)

	if l := len(aivs); l != 1 {
		t.Fatalf("Expected 1 element, the mean, but was %d", l)
	}

	expectedMean := 5.42
	mean := aivs[0]
	roundedMean := math.Floor(mean*100) / 100
	if roundedMean != expectedMean {
		t.Fatalf("Unexpected mean value for %f: rounded=%f, expected=%f", mean, roundedMean, expectedMean)
	}
}

func TestSampleInvocationsEmpty(t *testing.T) {
	ivs := []bench.Invocations{}

	aivs := bench.SampleInvocations(5)(ivs)
	if l := len(aivs); l != 0 {
		t.Fatalf("Expected 0 elements, got %d", l)
	}
}

func TestSampleInvocationsSome(t *testing.T) {
	ivs := []bench.Invocations{
		{Count: 5, Value: 4},
		{Count: 10, Value: 5},
		{Count: 20, Value: 6},
	}

	aivs := bench.SampleInvocations(5)(ivs)
	if l := len(aivs); l != 5 {
		t.Fatalf("Expected 5 elements, got %d", l)
	}
}

func TestSampleInvocationsAll(t *testing.T) {
	ivs := []bench.Invocations{
		{Count: 5, Value: 4},
		{Count: 10, Value: 5},
		{Count: 20, Value: 6},
	}

	aivs := bench.SampleInvocations(35)(ivs)
	checkAll(t, aivs)
}
