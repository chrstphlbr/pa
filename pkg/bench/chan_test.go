package bench_test

import (
	"fmt"
	"testing"

	"github.com/chrstphlbr/pa/pkg/bench"
)

const (
	invCount = 8
	invValue = 4.0
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
					Invocations: bench.Invocations{Count: invCount, Value: invValue},
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
				if inv != invValue {
					t.Fatalf("Expected invocation value %f, was %f", invValue, inv)
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

func TestTransformChan(t *testing.T) {
	factor := 2.0
	expCount := invCount
	expValue := invValue * factor

	checkExecution := func(e *bench.Execution) {
		for _, instance := range e.Instances {
			for _, trial := range instance.Trials {
				for _, fork := range trial.Forks {
					for _, iteration := range fork.Iterations {
						for _, invocation := range iteration.Invocations {
							if invocation.Count != expCount {
								t.Fatalf("invocation.Count not %d, was %d", expCount, invocation.Count)
							}
							if invocation.Value != expValue {
								t.Fatalf("unexpected invocation.Value: was %f, expected %f", invocation.Value, expValue)
							}
						}
					}
				}
			}
		}
	}

	expBenchs := 5
	bc := benchChan(1, expBenchs)
	bct := bench.TransformChan(
		bench.ConstantFactorExecutionTransformerFunc(factor, 0),
		bc,
	)

	var count int
	var started, ended bool
	for ev := range bct {
		switch ev.Type {
		case bench.ExecStart:
			started = true
		case bench.ExecEnd:
			ended = true
		case bench.ExecError:
			t.Fatalf("received an unexpected erronous ExecutionValues: %v", ev.Err)
		case bench.ExecNext:
			count++
			checkExecution(ev.Exec)
		}

		if ev.Err != nil {
			t.Fatalf("received an unexpected erronous ExecutionValues: %v", ev.Err)
		}
	}

	if !started {
		t.Fatalf("have not received an ExecutionValue of type bench.ExecStart")
	}

	if !ended {
		t.Fatalf("have not received an ExecutionValue of type bench.ExecEnd")
	}

	if count != expBenchs {
		t.Fatalf("did not get all benchmarks: expected %d, got %d", expBenchs, count)
	}
}
