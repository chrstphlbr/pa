package bench_test

import (
	"fmt"
	"testing"

	"bitbucket.org/sealuzh/pa/pkg/bench"
)

func benchChan(nr, elems int) bench.Chan {
	c := make(bench.Chan)

	go func() {
		defer close(c)
		c <- bench.ExecutionValue{
			Type: bench.ExecStart,
		}

		for i := 1; i <= elems; i++ {
			c <- bench.ExecutionValue{
				Type: bench.ExecNext,
				Exec: bench.NewExecutionFromInvocationsFlat(bench.InvocationsFlat{
					Benchmark:   bench.New(fmt.Sprintf("b%d", i)),
					Instance:    "i1",
					Trial:       nr,
					Fork:        0,
					Invocations: bench.Invocations{Count: 8, Value: 4},
				}),
			}
		}

		c <- bench.ExecutionValue{
			Type: bench.ExecEnd,
		}
	}()

	return c
}

func TestChanMerge(t *testing.T) {
	nrBenchs := 20

	c1 := benchChan(1, nrBenchs)
	c2 := benchChan(2, nrBenchs)
	c3 := benchChan(3, nrBenchs)
	mc := bench.MergeChans(c1, c2, c3)

	ev, ok := <-mc
	if !ok {
		t.Fatalf("No value received")
	}
	if ev.Type != bench.ExecStart {
		t.Fatalf("Expected bench.ExecStart")
	}

	for i := 1; i <= nrBenchs; i++ {
		ev, ok := <-mc
		if !ok {
			t.Fatalf("Expected benchmark %d", i)
		}

		if ev.Type != bench.ExecNext {
			t.Fatalf("Expected bench.ExecNext")
		}

		eb := bench.New(fmt.Sprintf("b%d", i))
		if !ev.Exec.Benchmark.Equals(eb) {
			t.Fatalf("Unexpected bench: %+v", eb)
		}

		sl := ev.Exec.Slice(bench.AllInvocations)
		if len(sl) != 1 {
			t.Fatalf("Exepected 1 instance")
		}

		if lts := len(sl[0]); lts != 3 {
			t.Fatalf("Expected 3 trials, was %d", lts)
		}

		for j := 0; j < 3; j++ {
			fs := sl[0][j]
			if len(fs) != 1 {
				t.Fatalf("Expected 1 fork for trial %d", j)
			}

			its := fs[0]
			if len(its) != 1 {
				t.Fatalf("Expected 1 iteration for trial %d", j)
			}

			for _, inv := range its[0] {
				if inv != 4 {
					t.Fatalf("Expected invocation value 4, was %f", inv)
				}
			}
		}
	}

	ev, ok = <-mc
	if !ok {
		t.Fatalf("No value received")
	}
	if ev.Type != bench.ExecEnd {
		t.Fatalf("Expected bench.ExecEnd")
	}
}
