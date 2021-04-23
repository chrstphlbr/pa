package bench_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/chrstphlbr/pa/pkg/bench"
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

func addInvocationsHelper(t *testing.T, e *bench.Execution, ivs []bench.InvocationsFlat) {
	for i, is := range ivs {
		err := e.AddInvocations(is)
		if err != nil {
			t.Fatalf("Could not add (pos=%d): %v", i, err)
		}
	}
}

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

func equalInstances(t *testing.T, i1, i2 *bench.Execution, failPointerEqual bool) {
	liids1 := len(i1.InstanceIDs)
	liids2 := len(i2.InstanceIDs)
	if liids1 != liids2 {
		t.Fatalf("InstanceIDs length not equal")
	}

	lis1 := len(i1.Instances)
	lis2 := len(i2.Instances)
	if lis1 != lis2 {
		t.Fatalf("Instances length not equal")
	}

	for i, instanceID1 := range i1.InstanceIDs {
		instanceID2 := i2.InstanceIDs[i]
		if instanceID1 != instanceID2 {
			t.Fatalf("InstanceID at position %d not equal", i)
		}

		instance1 := i1.Instances[instanceID1]
		instance2 := i2.Instances[instanceID2]

		if failPointerEqual && instance1 == instance2 {
			t.Fatalf("Instances are identical")
		}
		if instance1.ID != instance2.ID {
			t.Fatalf("Instance's IDs not equal")
		}
		if instance2.ID != instanceID2 {
			t.Fatalf("Instance's ID not equal to InstanceID")
		}

		equalTrials(t, instance1.TrialIDs, instance2.TrialIDs, instance1.Trials, instance2.Trials, failPointerEqual)
	}
}

func equalTrials(t *testing.T, trialIDs1, trialIDs2 []int, trials1, trials2 map[int]*bench.Trial, failPointerEqual bool) {
	if len(trialIDs1) != len(trialIDs2) {
		t.Fatalf("TrialIDs length not equal")
	}

	if len(trials1) != len(trials2) {
		t.Fatalf("Trials length not equal")
	}

	for i, trialID1 := range trialIDs1 {
		trialID2 := trialIDs2[i]
		if trialID1 != trialID2 {
			t.Fatalf("TrialID at position %d not equal", i)
		}

		trial1 := trials1[trialID1]
		trial2 := trials2[trialID2]

		if failPointerEqual && trial1 == trial2 {
			t.Fatalf("Trials are identical")
		}
		if trial1.ID != trial2.ID {
			t.Fatalf("Trials' IDs not equal")
		}
		if trial2.ID != trialID2 {
			t.Fatalf("Trial's ID not equal to TrialID")
		}

		equalForks(t, trial1.ForkIDs, trial2.ForkIDs, trial1.Forks, trial2.Forks, failPointerEqual)
	}
}

func equalForks(t *testing.T, forkIDs1, forkIDs2 []int, forks1, forks2 map[int]*bench.Fork, failPointerEqual bool) {
	if len(forkIDs1) != len(forkIDs2) {
		t.Fatalf("ForkIDs length not equal")
	}

	if len(forks1) != len(forks2) {
		t.Fatalf("Forks length not equal")
	}

	for i, forkID1 := range forkIDs1 {
		forkID2 := forkIDs2[i]
		if forkID1 != forkID2 {
			t.Fatalf("ForkID at position %d not equal", i)
		}

		fork1 := forks1[forkID1]
		fork2 := forks2[forkID2]

		if failPointerEqual && fork1 == fork2 {
			t.Fatalf("Forks are identical")
		}
		if fork1.ID != fork2.ID {
			t.Fatalf("Fork's IDs not equal")
		}
		if fork2.ID != forkID2 {
			t.Fatalf("Fork's ID not equal to ForkID")
		}

		equalIterations(t, fork1.IterationIDs, fork2.IterationIDs, fork1.Iterations, fork2.Iterations, failPointerEqual)
	}
}

func equalIterations(t *testing.T, iterationIDs1, iterationIDs2 []int, iterations1, iterations2 map[int]*bench.Iteration, failPointerEqual bool) {
	if len(iterationIDs1) != len(iterationIDs2) {
		t.Fatalf("IterationIDs length not equal")
	}

	if len(iterations1) != len(iterations2) {
		t.Fatalf("Iterations length not equal")
	}

	for i, iterationID1 := range iterationIDs1 {
		iterationID2 := iterationIDs2[i]
		if iterationID1 != iterationID2 {
			t.Fatalf("IterationID at position %d not equal", i)
		}

		iteration1 := iterations1[iterationID1]
		iteration2 := iterations2[iterationID2]

		if failPointerEqual && iteration1 == iteration2 {
			t.Fatalf("Iterations are identical")
		}
		if iteration1.ID != iteration2.ID {
			t.Fatalf("Iteration's IDs not equal")
		}
		if iteration2.ID != iterationID2 {
			t.Fatalf("Iteration's ID not equal to IterationID")
		}

		equalInvocations(t, iteration1.Invocations, iteration2.Invocations)
	}
}

func equalInvocations(t *testing.T, invocations1, invocations2 []bench.Invocations) {
	if len(invocations1) != len(invocations2) {
		t.Fatalf("Invocations length not equal")
	}
	for i, invocation1 := range invocations1 {
		invocation2 := invocations2[i]

		if invocation1.Count != invocation2.Count {
			t.Fatalf("Invocation.Counts not equal")
		}
		if invocation1.Value != invocation2.Value {
			t.Fatalf("Invocation.Values not equal")
		}
	}
}

// printing

func printExecution(e *bench.Execution) {
	fmt.Printf("\n== %+v ==\n", e.Benchmark)
	for _, insID := range e.InstanceIDs {
		ins := e.Instances[insID]
		fmt.Printf("- %s\n", ins.ID)
		for _, tID := range ins.TrialIDs {
			t := ins.Trials[tID]
			fmt.Printf(" - t%d\n", t.ID)
			for _, fID := range t.ForkIDs {
				f := t.Forks[fID]
				fmt.Printf("  - f%d\n", f.ID)
				for _, itID := range f.IterationIDs {
					it := f.Iterations[itID]
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
