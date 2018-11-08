package bootstrap_test

import (
	"fmt"
	"testing"

	"bitbucket.org/sealuzh/pa/pkg/bench"
	"bitbucket.org/sealuzh/pa/pkg/bootstrap"
	"bitbucket.org/sealuzh/pa/pkg/stat"
)

func TestCIsEmpty(t *testing.T) {
	bc := make(bench.Chan)
	close(bc)

	rc := bootstrap.CIs(bc, 1, 1, stat.Mean, 0.05)

	_, ok := <-rc
	if ok {
		t.Fatalf("Result channel has values")
	}
}

func TestCIsNoValues(t *testing.T) {
	bc := make(bench.Chan)
	go func() {
		defer close(bc)
		bc <- bench.ExecutionValue{
			Type: bench.ExecStart,
		}
		bc <- bench.ExecutionValue{
			Type: bench.ExecEnd,
		}
	}()

	rc := bootstrap.CIs(bc, 1, 1, stat.Mean, 0.05)

	_, ok := <-rc
	if ok {
		t.Fatalf("Result channel has values")
	}
}

func TestCIsError(t *testing.T) {
	bc := make(bench.Chan)

	sendErr := fmt.Errorf("")

	go func() {
		defer close(bc)
		bc <- bench.ExecutionValue{
			Type: bench.ExecStart,
		}

		bc <- bench.ExecutionValue{
			Type: bench.ExecError,
			Err:  sendErr,
		}

		bc <- bench.ExecutionValue{
			Type: bench.ExecEnd,
		}
	}()

	rc := bootstrap.CIs(bc, 1, 1, stat.Mean, 0.05)

	ev, ok := <-rc
	if !ok {
		t.Fatalf("Expected error, but no elements sent")
	}
	if ev.Err != sendErr {
		t.Fatalf("Unexepected error received: '%v'", ev.Err)
	}

	_, ok = <-rc
	if ok {
		t.Fatalf("Result channel has values")
	}
}

func createChannel(from, to int) (bench.Chan, []*bench.Execution) {
	bc := make(bench.Chan)

	var execs []*bench.Execution
	for i := from; i < to; i++ {
		b := bench.New(fmt.Sprintf("b%d", i))

		iid := bench.NewInstanceID("i1")
		execs = append(execs, &bench.Execution{
			Benchmark: b,
			Instances: map[bench.InstanceID]*bench.Instance{
				iid: {
					ID: iid,
					Trials: []bench.Trial{
						[]bench.Fork{
							[]bench.Iteration{
								bench.Iteration(bench.Invocations{4.0, 4.0, 4.0, 4.0, 4.0}),
							},
						},
					},
				},
			},
		})
	}

	go func() {
		defer close(bc)
		bc <- bench.ExecutionValue{
			Type: bench.ExecStart,
		}

		for _, e := range execs {
			bc <- bench.ExecutionValue{
				Type: bench.ExecNext,
				Exec: e,
			}
		}

		bc <- bench.ExecutionValue{
			Type: bench.ExecEnd,
		}
	}()
	return bc, execs
}

func TestCIsValues(t *testing.T) {
	bc, execs := createChannel(0, 10)

	rc := bootstrap.CIs(bc, 2, 1, stat.Mean, 0.05)

	for i, e := range execs {
		ev, ok := <-rc
		if !ok {
			t.Fatalf("Expected value from channel, but did not receive one (pos: %d, bench: %v)", i, e.Benchmark)
		}
		if ev.Err != nil {
			t.Fatalf("Received error: %v", ev.Err)
		}
		if !ev.Benchmark.Equals(e.Benchmark) {
			t.Fatalf("Expected benchmark %v, got %v", e, ev)
		}
		if ev.CI.Level != 0.95 {
			t.Fatalf("Unexpected CI level %.2f", ev.CI.Level)
		}
		if ev.CI.Lower != 4 {
			t.Fatalf("Unexpected CI lower %.2f", ev.CI.Lower)
		}
		if ev.CI.Upper != 4 {
			t.Fatalf("Unexpected CI upper %.2f", ev.CI.Upper)
		}
	}

	_, ok := <-rc
	if ok {
		t.Fatalf("Result channel has values")
	}
}