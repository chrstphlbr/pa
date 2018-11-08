package bench_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"bitbucket.org/sealuzh/pa/pkg/bench"
)

func TestBNewEmpty(t *testing.T) {
	b := bench.New("b")
	if b.Name != "b" {
		t.Fatalf("Invalid B.Name")
	}
	if len(b.FunctionParams) != 0 {
		t.Fatalf("FunctionParams not 0")
	}
	if len(b.PerfParamKeys()) != 0 {
		t.Fatalf("PerfParamKeys not 0")
	}
	if len(b.PerfParams()) != 0 {
		t.Fatalf("PerfParams not 0")
	}
}

func TestBAddPerfParam(t *testing.T) {
	b := bench.New("b")
	for i := 0; i < 10; i++ {
		b.AddPerfParam(fmt.Sprintf("p%d", i), fmt.Sprintf("v%d", i))
		el := i + 1

		// integrety checks
		if b.Name != "b" {
			t.Fatalf("Name has changed from 'b' to '%s'", b.Name)
		}
		if lfp := len(b.FunctionParams); lfp != 0 {
			t.Fatalf("FunctionParams length has changed from %d to %d", 0, lfp)
		}

		// check keys and values
		keys := b.PerfParamKeys()
		lppk := len(keys)
		if lppk != el {
			t.Fatalf("PerfParamKeys length invalid: expected %d, was %d", el, lppk)
		}

		params := b.PerfParams()
		lpp := len(params)
		if lpp != el {
			t.Fatalf("PerfParams length invalid: expected %d, was %d", el, lppk)
		}

		for j := 0; j <= i; j++ {
			key := keys[j]
			if e := fmt.Sprintf("p%d", j); key != e {
				t.Fatalf("Invalid key: expected '%s', was '%s'", e, key)
			}

			v, ok := params[key]
			if !ok {
				t.Fatalf("PerfPram (key = '%s') not in benchmark", key)
			}

			if e := fmt.Sprintf("v%d", j); v != e {
				t.Fatalf("Invalid value for '%s': expected '%s', was '%s'", key, e, v)
			}
		}
	}
}

func TestBAddPerfParamSorted(t *testing.T) {
	b := bench.New("b")

	random := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	// create random order of elements
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for n := len(random); n > 0; n-- {
		randIndex := r.Intn(n)
		random[n-1], random[randIndex] = random[randIndex], random[n-1]
	}

	for _, v := range random {
		b.AddPerfParam(fmt.Sprintf("p%d", v), fmt.Sprintf("v%d", v))
	}

	keys := b.PerfParamKeys()
	params := b.PerfParams()
	for i, key := range keys {
		// check key
		if e := fmt.Sprintf("p%d", i); key != e {
			t.Fatalf("Invalid key: expected '%s', was '%s'", e, key)
		}
		// check value
		v := params[key]
		if e := fmt.Sprintf("v%d", i); v != e {
			t.Fatalf("Invalid value: expected '%s', was '%s'", e, v)
		}
	}
}
