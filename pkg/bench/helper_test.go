package bench_test

import (
	"fmt"
	"math/rand"
	"testing"

	"bitbucket.org/sealuzh/pa/pkg/bench"
)

// creating

type invGen func(int) []bench.Invocations

func createRandomInvocations(count int) []bench.Invocations {
	const (
		min = 0.0
		max = 10000.0
	)
	is := make([]bench.Invocations, count)
	for i := 0; i < count; i++ {
		r := min + rand.Float64()*(max-min)
		is[i] = bench.Invocations{Count: 1, Value: r}
	}
	return is
}

func createInvocations(count int) []bench.Invocations {
	is := make([]bench.Invocations, count)
	for i := 0; i < count; i++ {
		is[i] = bench.Invocations{Count: 1, Value: float64(i)}
	}
	return is
}

func createInvocationsFlatGenerator(count int, b *bench.B, ins string, t, f, i int, invGen invGen) []bench.InvocationsFlat {
	is := invGen(count)

	ivs := make([]bench.InvocationsFlat, len(is))
	for j, iv := range is {
		ivs[j] = bench.InvocationsFlat{
			Benchmark:   b,
			Instance:    ins,
			Trial:       t,
			Fork:        f,
			Iteration:   i,
			Invocations: iv,
		}
	}
	return ivs
}

func createInvocationsFlat(count int, b *bench.B, ins string, t, f, i int) []bench.InvocationsFlat {
	return createInvocationsFlatGenerator(count, b, ins, t, f, i, createInvocations)
}

// func createRandomInvocationsFlat(count int, b *bench.B, ins string, t, f, i int) bench.InvocationsFlat {
// 	return createInvocationsFlatGenerator(count, b, ins, t, f, i, createRandomInvocations)
// }

// checking

func checkInstance(t *testing.T, e *bench.Execution, in string, nrivs int, enrins, enrt, enrf, enri int, enrinvs int) {
	if lins := len(e.Instances); lins != enrins {
		t.Fatalf("Instances length not %d, was %d", enrins, lins)
	}
	if lins := len(e.InstanceIDs); lins != enrins {
		t.Fatalf("InstanceIDs length not %d, was %d", enrins, lins)
	}

	iid := in
	i, ie := e.Instances[iid]
	if !ie {
		t.Fatalf("Instance '%s' does not exist", iid)
	}

	// check InstanceIds
	var contained bool
	for _, i := range e.InstanceIDs {
		if i == iid {
			contained = true
			break
		}
	}
	if !contained {
		t.Fatalf("Instance ID '%s' not in InstanceIDs", iid)
	}

	if lts := len(i.Trials); lts != enrt {
		t.Fatalf("Trials length not %d, was %d", enrt, lts)
	}
	if lts := len(i.TrialIDs); lts != enrt {
		t.Fatalf("TrialIDs length not %d, was %d", enrt, lts)
	}

	for trial := 0; trial < enrt; trial++ {
		tid := i.TrialIDs[trial]
		if tid != trial+1 {
			t.Fatalf("Invalid TrialID: expected %d, was %d", trial, tid)
		}
		tr, ok := i.Trials[tid]
		if !ok {
			t.Fatalf("No Trial with id %d", tid)
		}
		checkTrial(t, tr, nrivs, enrf, enri, enrinvs)
	}
}

func checkTrial(t *testing.T, tr *bench.Trial, nrivs int, enrf, enri int, enrinvs int) {
	if lfs := len(tr.Forks); lfs != enrf {
		t.Fatalf("Forks length not %d, was %d", enrf, lfs)
	}
	if lfs := len(tr.ForkIDs); lfs != enrf {
		t.Fatalf("ForkIDs length not %d, was %d", enrf, lfs)
	}

	for fork := 0; fork < enrf; fork++ {
		fid := tr.ForkIDs[fork]
		if fid != fork+1 {
			t.Fatalf("Invalid ForkID: expected %d, was %d", fork, fid)
		}
		f, ok := tr.Forks[fid]
		if !ok {
			t.Fatalf("No Fork with id %d", fid)
		}
		checkFork(t, f, nrivs, enri, enrinvs)
	}
}

func checkFork(t *testing.T, f *bench.Fork, nrivs int, enri int, enrinvs int) {
	if lis := len(f.Iterations); lis != enri {
		t.Fatalf("Iterations length not %d, was %d", enri, lis)
	}
	if lis := len(f.IterationIDs); lis != enri {
		t.Fatalf("IterationIDs length not %d, was %d", enri, lis)
	}

	for iter := 0; iter < enri; iter++ {
		id := f.IterationIDs[iter]
		if id != iter+1 {
			t.Fatalf("Invalid IterationID: expected %d, was %d", iter, id)
		}
		i, ok := f.Iterations[id]
		if !ok {
			t.Fatalf("No Iteration with id %d", id)
		}
		checkIteration(t, i, nrivs, enrinvs)
	}
}

func checkIteration(t *testing.T, i *bench.Iteration, nrivs int, enrinvs int) {
	if lis := len(i.Invocations); lis != enrinvs {
		t.Fatalf("Invocations length not %d, was %d", enrinvs, lis)
	}

	for i, v := range i.Invocations {
		expected := float64(i % nrivs)
		if v.Count != 1 {
			t.Fatalf("Expected invocation count to be %d, was %d", 1, v.Count)
		}
		if v.Value != expected {
			t.Fatalf("Unexepected invocations value %f, expected %f", v.Value, expected)
		}
	}
}

// printing

func printExecution(e *bench.Execution) {
	fmt.Printf("\n== %+v ==\n", e.Benchmark)
	for _, ins := range e.Instances {
		fmt.Printf("- %s\n", ins.ID)
		for _, t := range ins.Trials {
			fmt.Printf(" - t%d\n", t.ID)
			for _, f := range t.Forks {
				fmt.Printf("  - f%d\n", f.ID)
				for _, it := range f.Iterations {
					fmt.Printf("   - i%d\n", it.ID)
					for _, iv := range it.Invocations {
						fmt.Printf("    - %+v\n", iv)
					}
				}
			}
		}
	}
}

func printExecutions(es []*bench.Execution) {
	for _, e := range es {
		printExecution(e)
	}
}
