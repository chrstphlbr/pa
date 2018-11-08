package bootstrap_test

import (
	"fmt"
	"testing"

	"bitbucket.org/sealuzh/pa/pkg/stat"

	"bitbucket.org/sealuzh/pa/pkg/bench"
	"bitbucket.org/sealuzh/pa/pkg/bootstrap"
)

func checkChannelEmpty(t *testing.T, rc <-chan bootstrap.CIRatioResult) {
	_, ok := <-rc
	if ok {
		t.Fatalf("Result channel has values")
	}
}
func TestCIRatiosEmpty(t *testing.T) {
	bc1 := make(bench.Chan)
	close(bc1)
	bc2 := make(bench.Chan)
	close(bc2)

	rc := bootstrap.CIRatios(bc1, bc2, 1, 1, stat.Mean, 0.05)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosNoValues(t *testing.T) {
	bc1 := make(bench.Chan)
	go func() {
		defer close(bc1)
		bc1 <- bench.ExecutionValue{
			Type: bench.ExecStart,
		}
		bc1 <- bench.ExecutionValue{
			Type: bench.ExecEnd,
		}
	}()

	bc2 := make(bench.Chan)
	go func() {
		defer close(bc2)
		bc2 <- bench.ExecutionValue{
			Type: bench.ExecStart,
		}
		bc2 <- bench.ExecutionValue{
			Type: bench.ExecEnd,
		}
	}()

	rc := bootstrap.CIRatios(bc1, bc2, 1, 1, stat.Mean, 0.05)

	checkChannelEmpty(t, rc)
}

func ciRatiosError(t *testing.T, side int) {
	if side != 1 && side != 2 {
		t.Fatalf("Invalid side %d", side)
	}

	bc1 := make(bench.Chan)

	sendErr := fmt.Errorf("")

	go func() {
		defer close(bc1)
		bc1 <- bench.ExecutionValue{
			Type: bench.ExecStart,
		}

		bc1 <- bench.ExecutionValue{
			Type: bench.ExecError,
			Err:  sendErr,
		}

		bc1 <- bench.ExecutionValue{
			Type: bench.ExecEnd,
		}
	}()

	nrExecs := 5
	bc2, execs := createChannel(0, nrExecs)

	var rc <-chan bootstrap.CIRatioResult
	if side == 1 {
		rc = bootstrap.CIRatios(bc1, bc2, 2, 1, stat.Mean, 0.05)
	} else if side == 2 {
		rc = bootstrap.CIRatios(bc2, bc1, 2, 1, stat.Mean, 0.05)
	}

	ev, ok := <-rc
	if !ok {
		t.Fatalf("Expected error, but no elements sent")
	}
	if ev.Err != sendErr {
		t.Fatalf("Unexepected error received: '%v'", ev.Err)
	}

	if side == 1 {
		checkOneSided(t, rc, execs, 0, nrExecs, 0, 2)
	} else if side == 2 {
		checkOneSided(t, rc, execs, 0, nrExecs, 0, 1)
	}

	checkChannelEmpty(t, rc)
}

func checkOneSided(t *testing.T, rc <-chan bootstrap.CIRatioResult, execs []*bench.Execution, from, to, add, side int) {
	for i := from; i < to; i++ {
		e := execs[i+add]
		ev, ok := <-rc
		if !ok {
			t.Fatalf("Expected value from channel, but did not receive one (pos: %d, bench: %v)", i, e.Benchmark)
		}
		if ev.Err != nil {
			t.Fatalf("Received error: %v", ev.Err)
		}
		if !ev.Benchmark.Equals(e.Benchmark) {
			t.Fatalf("Expected benchmark %v, got %v", e.Benchmark, ev.Benchmark)
		}

		eci := stat.CI{
			Level: 0.95,
			Lower: 4,
			Upper: 4,
		}

		var ecir stat.CIRatio
		if side == 1 {
			// right channel has values
			ecir = stat.CIRatio{
				CIA: eci,
			}
		} else if side == 2 {
			// left channel has values
			ecir = stat.CIRatio{
				CIB: eci,
			}
		}

		if ev.CIRatio != ecir {
			t.Fatalf("Unexpected CIRation (pos: %d): was %+v, expected %+v", i, ev.CIRatio, ecir)
		}
	}
}

func TestCIRatiosError1(t *testing.T) {
	ciRatiosError(t, 1)
}

func TestCIRatiosError2(t *testing.T) {
	ciRatiosError(t, 2)
}

func checkMerged(t *testing.T, rc <-chan bootstrap.CIRatioResult, ex1, ex2 []*bench.Execution, from, to, i1Add, i2Add int) {
	for i := from; i < to; i++ {
		e1 := ex1[i+i1Add]
		e2 := ex2[i+i2Add]
		ev, ok := <-rc
		if !ok {
			t.Fatalf("Expected value from channel, but did not receive one (pos: %d)", i)
		}
		if ev.Err != nil {
			t.Fatalf("Received error (pos: %d): %v", i, ev.Err)
		}
		if !ev.Benchmark.Equals(e1.Benchmark) || !ev.Benchmark.Equals(e2.Benchmark) {
			t.Fatalf("Expected benchmarks %v, %v; got %v (pos: %d)", e1.Benchmark, e2.Benchmark, ev.Benchmark, i)
		}

		eci := stat.CI{
			Level: 0.95,
			Lower: 4,
			Upper: 4,
		}

		ecir := stat.CIRatio{
			CIA: eci,
			CIB: eci,
			CIRatio: stat.CI{
				Level: 0.95,
				Lower: 1,
				Upper: 1,
			},
		}

		if ev.CIRatio != ecir {
			t.Fatalf("Unexpected CIRation (pos: %d): was %+v, expected %+v", i, ev.CIRatio, ecir)
		}
	}
}

func TestCIRatios1empty(t *testing.T) {
	bc1, ex1 := createChannel(0, 10)
	bc2 := make(bench.Chan)
	close(bc2)

	rc := bootstrap.CIRatios(bc1, bc2, 2, 1, stat.Mean, 0.05)

	// sole channel 2
	checkOneSided(t, rc, ex1, 0, 10, 0, 1)

	checkChannelEmpty(t, rc)
}

func TestCIRatios2empty(t *testing.T) {
	bc1 := make(bench.Chan)
	close(bc1)
	bc2, ex2 := createChannel(0, 10)

	rc := bootstrap.CIRatios(bc1, bc2, 2, 1, stat.Mean, 0.05)

	// sole channel 2
	checkOneSided(t, rc, ex2, 0, 10, 0, 2)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValues1earlier(t *testing.T) {
	bc1, ex1 := createChannel(0, 7)
	bc2, ex2 := createChannel(0, 10)

	rc := bootstrap.CIRatios(bc1, bc2, 2, 1, stat.Mean, 0.05)

	// check ratios
	checkMerged(t, rc, ex1, ex2, 0, 7, 0, 0)

	// sole channel 2
	checkOneSided(t, rc, ex2, 7, 10, 0, 2)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValues2earlier(t *testing.T) {
	bc1, ex1 := createChannel(0, 10)
	bc2, ex2 := createChannel(0, 7)

	rc := bootstrap.CIRatios(bc1, bc2, 2, 1, stat.Mean, 0.05)

	// check ratios
	checkMerged(t, rc, ex1, ex2, 0, 7, 0, 0)

	// sole channel 2
	checkOneSided(t, rc, ex1, 7, 10, 0, 1)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValues1later(t *testing.T) {
	bc1, ex1 := createChannel(3, 10)
	bc2, ex2 := createChannel(0, 10)

	rc := bootstrap.CIRatios(bc1, bc2, 2, 1, stat.Mean, 0.05)

	// sole channel
	checkOneSided(t, rc, ex2, 0, 3, 0, 2)

	// check ratios
	checkMerged(t, rc, ex1, ex2, 3, 10, -3, 0)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValues2later(t *testing.T) {
	bc1, ex1 := createChannel(0, 10)
	bc2, ex2 := createChannel(3, 10)

	rc := bootstrap.CIRatios(bc1, bc2, 2, 1, stat.Mean, 0.05)

	// sole channel
	checkOneSided(t, rc, ex1, 0, 3, 0, 1)

	// check ratios
	checkMerged(t, rc, ex1, ex2, 3, 10, 0, -3)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValues1middle(t *testing.T) {
	bc1, ex1 := createChannel(3, 7)
	bc2, ex2 := createChannel(0, 10)

	rc := bootstrap.CIRatios(bc1, bc2, 2, 1, stat.Mean, 0.05)

	// sole channel 2 start
	checkOneSided(t, rc, ex2, 0, 3, 0, 2)

	// check ratios
	checkMerged(t, rc, ex1, ex2, 3, 7, -3, 0)

	// sole channel 2 end
	checkOneSided(t, rc, ex2, 7, 10, 0, 2)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValues2middle(t *testing.T) {
	bc1, ex1 := createChannel(0, 10)
	bc2, ex2 := createChannel(3, 7)

	rc := bootstrap.CIRatios(bc1, bc2, 2, 1, stat.Mean, 0.05)

	// sole channel 1 start
	checkOneSided(t, rc, ex1, 0, 3, 0, 1)

	// check ratios
	checkMerged(t, rc, ex1, ex2, 3, 7, 0, -3)

	// sole channel 1 end
	checkOneSided(t, rc, ex1, 7, 10, 0, 1)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValues(t *testing.T) {
	bc1, ex1 := createChannel(0, 7)
	bc2, ex2 := createChannel(5, 10)

	rc := bootstrap.CIRatios(bc1, bc2, 2, 1, stat.Mean, 0.05)

	// sole channel 1
	checkOneSided(t, rc, ex1, 0, 5, 0, 1)

	// check ratios
	checkMerged(t, rc, ex1, ex2, 5, 7, 0, -5)

	// sole channel 2
	checkOneSided(t, rc, ex2, 7, 10, -5, 2)

	checkChannelEmpty(t, rc)
}
