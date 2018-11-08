package bench_test

import (
	"fmt"
	"math/rand"
	"testing"

	"bitbucket.org/sealuzh/pa/pkg/bench"
)

// creating

type invGen func(int) bench.Invocations

func createRandomInvocations(count int) bench.Invocations {
	const (
		min = 0.0
		max = 10000.0
	)

	is := bench.Invocations{}
	for i := 0; i < 10; i++ {
		r := min + rand.Float64()*(max-min)
		is = append(is, r)
	}
	return is
}

func createInvocations(count int) bench.Invocations {
	is := bench.Invocations{}
	for i := 0; i < count; i++ {
		is = append(is, float64(i))
	}
	return is
}

func createInvocationsFlatGenerator(count int, b *bench.B, ins string, t, f, i int, invGen invGen) bench.InvocationsFlat {
	is := invGen(count)

	return bench.InvocationsFlat{
		Benchmark:   b,
		Instance:    ins,
		Trial:       t,
		Fork:        f,
		Iteration:   i,
		Invocations: is,
	}
}

func createInvocationsFlat(count int, b *bench.B, ins string, t, f, i int) bench.InvocationsFlat {
	return createInvocationsFlatGenerator(count, b, ins, t, f, i, createInvocations)
}

func createRandomInvocationsFlat(count int, b *bench.B, ins string, t, f, i int) bench.InvocationsFlat {
	return createInvocationsFlatGenerator(count, b, ins, t, f, i, createRandomInvocations)
}

// checking

func checkInstance(t *testing.T, e *bench.Execution, in string, nrivs int, enrins, enrt, enrf, enri int, enrinvs int) {
	iid := bench.NewInstanceID(in)
	i, ie := e.Instances[iid]
	if !ie {
		t.Fatalf("Instance '%s' does not exist", in)
	}

	lins := len(e.Instances)
	if lins != enrins {
		t.Fatalf("Instance length not %d, was %d", enrins, lins)
	}

	for trial := 0; trial < enrt; trial++ {
		checkTrials(t, i, nrivs, trial, enrt, enrf, enri, enrinvs)
	}
}

func checkTrials(t *testing.T, i *bench.Instance, nrivs int, trial int, enrt, enrf, enri int, enrinvs int) {
	lt := len(i.Trials)
	if lt != enrt {
		t.Fatalf("Trial length not %d, was %d", enrt, lt)
	}

	for fork := 0; fork < enrf; fork++ {
		checkFork(t, i, nrivs, trial, fork, enrf, enri, enrinvs)
	}
}

func checkFork(t *testing.T, i *bench.Instance, nrivs int, trial int, fork int, enrf, enri int, enrinvs int) {
	forks := i.Trials[trial]
	lf := len(forks)
	if lf != enrf {
		t.Fatalf("Forks length not %d, was %d", enrf, lf)
	}

	for iter := 0; iter < enri; iter++ {
		checkIter(t, i, nrivs, trial, fork, iter, enri, enrinvs)
	}
}

func checkIter(t *testing.T, i *bench.Instance, nrivs int, trial int, fork int, iter int, enri int, enrinvs int) {
	iterations := i.Trials[trial][fork]
	li := len(iterations)
	if li != enri {
		t.Fatalf("Iterations length not %d, was %d", enri, li)
	}

	invocations := i.Trials[trial][fork][iter]
	livs := len(invocations)
	if livs != enrinvs {
		t.Fatalf("Invocations length not %d, was %d", enrinvs, livs)
	}

	for i, v := range invocations {
		expected := float64(i % nrivs)
		if v != expected {
			t.Fatalf("Unexepected invocations value %f, expected %f", v, expected)
		}
	}
}

// printing

func printExecution(e *bench.Execution) {
	for _, ins := range e.Instances {
		fmt.Printf("- %s\n", ins.ID)
		for t, fs := range ins.Trials {
			fmt.Printf(" - t%d\n", t)
			for f, is := range fs {
				fmt.Printf("  - f%d\n", f)
				for i, ivs := range is {
					fmt.Printf("   - i%d\n", i)
					for _, iv := range ivs {
						fmt.Printf("    - %f\n", iv)
					}
				}
			}
		}
	}
}

func printExecutions(es []*bench.Execution) {
	for _, e := range es {
		fmt.Printf("\n== %+v ==\n", e.Benchmark)
		printExecution(e)
	}
}
