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

	cif, _ := ciFuncs(1, 1, stat.Mean, []float64{0.05, 0.01}, bench.AllInvocations)
	rc := bootstrap.CIs(bc, cif)

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

	cif, _ := ciFuncs(1, 1, stat.Mean, []float64{0.05, 0.01}, bench.AllInvocations)
	rc := bootstrap.CIs(bc, cif)

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

	cif, _ := ciFuncs(1, 1, stat.Mean, []float64{0.05, 0.01}, bench.AllInvocations)
	rc := bootstrap.CIs(bc, cif)

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
	return createChannelStartEnd(from, to, true, true)
}

func createChannelStartEnd(from, to int, start, end bool) (bench.Chan, []*bench.Execution) {
	return createChannelStartEndDetail(from, to, start, end, "i1", 0, 0, 0)
}

func createChannelStartEndDetail(from, to int, start, end bool, instance string, trial, fork, iteration int) (bench.Chan, []*bench.Execution) {
	bc := make(bench.Chan)

	var execs []*bench.Execution
	for i := from; i < to; i++ {
		b := bench.New(fmt.Sprintf("b%d", i))
		e := bench.NewExecutionFromInvocationsFlat(bench.InvocationsFlat{
			Benchmark:   b,
			Instance:    instance,
			Trial:       trial,
			Fork:        fork,
			Iteration:   iteration,
			Invocations: bench.Invocations{Count: 5, Value: 4.0},
		})
		execs = append(execs, e)
	}

	go func() {
		defer close(bc)
		if start {
			bc <- bench.ExecutionValue{
				Type: bench.ExecStart,
			}
		}

		for _, e := range execs {
			bc <- bench.ExecutionValue{
				Type: bench.ExecNext,
				Exec: e,
			}
		}

		if end {
			bc <- bench.ExecutionValue{
				Type: bench.ExecEnd,
			}
		}
	}()
	return bc, execs
}

func TestCIsValues(t *testing.T) {
	bc, execs := createChannel(0, 10)

	sigLevels := []float64{0.05, 0.01}

	cif, _ := ciFuncs(2, 1, stat.Mean, sigLevels, bench.AllInvocations)
	rc := bootstrap.CIs(bc, cif)

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

		for i, sigLevel := range sigLevels {
			ciLevel := 1 - sigLevel
			ci := ev.CIs[i]
			if ci.Level != ciLevel {
				t.Fatalf("Unexpected CI level: expected %.2f but was %.2f", ciLevel, ci.Level)
			}
			if ci.Lower != 4 {
				t.Fatalf("Unexpected CI lower %.2f", ci.Lower)
			}
			if ci.Upper != 4 {
				t.Fatalf("Unexpected CI upper %.2f", ci.Upper)
			}
		}
	}

	_, ok := <-rc
	if ok {
		t.Fatalf("Result channel has values")
	}
}
