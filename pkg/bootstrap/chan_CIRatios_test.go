package bootstrap_test

import (
	"fmt"
	"testing"

	"github.com/chrstphlbr/pa/pkg/stat"
	"github.com/chrstphlbr/pa/pkg/bench"
	"github.com/chrstphlbr/pa/pkg/bootstrap"
)

var ciLevels = []float64{0.05, 0.01}

func checkChannelEmpty(t *testing.T, rc <-chan bootstrap.CIRatioResult) {
	_, ok := <-rc
	if ok {
		t.Fatalf("Result channel has values")
	}
}

func ciFuncs(sim, nrWorkers int, sf stat.StatisticFunc, sls []float64, sampler bench.InvocationSampler) (bootstrap.CIFunc, bootstrap.CIRatioFunc) {
	return bootstrap.CIFuncSetup(sim, nrWorkers, sf, sls, sampler), bootstrap.CIRatioFuncSetup(sim, nrWorkers, sf, sls, sampler)
}
func TestCIRatiosEmpty(t *testing.T) {
	bc1 := make(bench.Chan)
	close(bc1)
	bc2 := make(bench.Chan)
	close(bc2)

	cif, cirf := ciFuncs(1, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

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

	cif, cirf := ciFuncs(1, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

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
	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	if side == 1 {
		rc = bootstrap.CIRatios(bc1, bc2, cif, cirf)
	} else if side == 2 {
		rc = bootstrap.CIRatios(bc2, bc1, cif, cirf)
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

func checkRatios(t *testing.T, ratioPos int, cirs, ecirs []stat.CIRatio) {
	ecirsLen := len(ecirs)
	if cirsLen := len(cirs); ecirsLen != cirsLen {
		t.Fatalf("Unexpected CIRations length (pos: %d): was %d, expected %d", ratioPos, cirsLen, ecirsLen)
	}

	for i, ecir := range ecirs {
		cir := cirs[i]

		if cir != ecir {
			t.Fatalf("Unexpected CIRatio (pos: %d): was %+v, expected %+v", i, cir, ecir)
		}
	}
}

func sidedEcis(side int, ecis []stat.CI) []stat.CIRatio {
	ecirs := make([]stat.CIRatio, len(ecis))
	if side == 1 {
		// right channel has values
		for i, eci := range ecis {
			ecirs[i] = stat.CIRatio{
				CIA: eci,
			}
		}
	} else if side == 2 {
		// left channel has values
		for i, eci := range ecis {
			ecirs[i] = stat.CIRatio{
				CIB: eci,
			}
		}
	}
	return ecirs
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

		ecis := []stat.CI{
			stat.CI{
				Metric: 4,
				Level:  0.95,
				Lower:  4,
				Upper:  4,
			},
			stat.CI{
				Metric: 4,
				Level:  0.99,
				Lower:  4,
				Upper:  4,
			},
		}

		ecirs := sidedEcis(side, ecis)

		checkRatios(t, i, ev.CIRatios, ecirs)
	}
}

func TestCIRatiosError1(t *testing.T) {
	ciRatiosError(t, 1)
}

func TestCIRatiosError2(t *testing.T) {
	ciRatiosError(t, 2)
}

func mergedEcis(ecis, eciRatios []stat.CI) []stat.CIRatio {
	ecirs := make([]stat.CIRatio, len(ecis))
	for i, eci := range ecis {
		ecirs[i] = stat.CIRatio{
			CIA:     eci,
			CIB:     eci,
			CIRatio: eciRatios[i],
		}
	}
	return ecirs
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

		ecis := []stat.CI{
			stat.CI{
				Metric: 4,
				Level:  0.95,
				Lower:  4,
				Upper:  4,
			},
			stat.CI{
				Metric: 4,
				Level:  0.99,
				Lower:  4,
				Upper:  4,
			},
		}

		eciRatios := []stat.CI{
			stat.CI{
				Metric: 1,
				Level:  0.95,
				Lower:  1,
				Upper:  1,
			},
			stat.CI{
				Metric: 1,
				Level:  0.99,
				Lower:  1,
				Upper:  1,
			},
		}

		eciCIRatios := mergedEcis(ecis, eciRatios)

		checkRatios(t, i, ev.CIRatios, eciCIRatios)
	}
}

func TestCIRatios1empty(t *testing.T) {
	bc1, ex1 := createChannel(0, 10)
	bc2 := make(bench.Chan)
	close(bc2)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// sole channel 2
	checkOneSided(t, rc, ex1, 0, 10, 0, 1)

	checkChannelEmpty(t, rc)
}

func TestCIRatios2empty(t *testing.T) {
	bc1 := make(bench.Chan)
	close(bc1)
	bc2, ex2 := createChannel(0, 10)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// sole channel 2
	checkOneSided(t, rc, ex2, 0, 10, 0, 2)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValues1earlier(t *testing.T) {
	bc1, ex1 := createChannel(0, 7)
	bc2, ex2 := createChannel(0, 10)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// check ratios
	checkMerged(t, rc, ex1, ex2, 0, 7, 0, 0)

	// sole channel 2
	checkOneSided(t, rc, ex2, 7, 10, 0, 2)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValues2earlier(t *testing.T) {
	bc1, ex1 := createChannel(0, 10)
	bc2, ex2 := createChannel(0, 7)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// check ratios
	checkMerged(t, rc, ex1, ex2, 0, 7, 0, 0)

	// sole channel 2
	checkOneSided(t, rc, ex1, 7, 10, 0, 1)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValues1later(t *testing.T) {
	bc1, ex1 := createChannel(3, 10)
	bc2, ex2 := createChannel(0, 10)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// sole channel
	checkOneSided(t, rc, ex2, 0, 3, 0, 2)

	// check ratios
	checkMerged(t, rc, ex1, ex2, 3, 10, -3, 0)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValues2later(t *testing.T) {
	bc1, ex1 := createChannel(0, 10)
	bc2, ex2 := createChannel(3, 10)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// sole channel
	checkOneSided(t, rc, ex1, 0, 3, 0, 1)

	// check ratios
	checkMerged(t, rc, ex1, ex2, 3, 10, 0, -3)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValues1middle(t *testing.T) {
	bc1, ex1 := createChannel(3, 7)
	bc2, ex2 := createChannel(0, 10)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

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

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

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

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// sole channel 1
	checkOneSided(t, rc, ex1, 0, 5, 0, 1)

	// check ratios
	checkMerged(t, rc, ex1, ex2, 5, 7, 0, -5)

	// sole channel 2
	checkOneSided(t, rc, ex2, 7, 10, -5, 2)

	checkChannelEmpty(t, rc)
}

func appendChannels(cs ...bench.Chan) bench.Chan {
	out := make(bench.Chan)

	go func() {
		defer close(out)
		for _, c := range cs {
			for v := range c {
				out <- v
			}
		}
	}()

	return out
}

func TestCIRatiosValuesGap1(t *testing.T) {
	bc1, ex1 := createChannelStartEnd(2, 5, true, true)
	bc21, ex21 := createChannelStartEnd(0, 2, true, false)
	bc22, ex22 := createChannelStartEnd(5, 10, false, true)

	bc2 := appendChannels(bc21, bc22)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// sole channel 2
	checkOneSided(t, rc, ex21, 0, 2, 0, 2)

	// sole channel 1
	checkOneSided(t, rc, ex1, 2, 5, -2, 1)

	// sole channel 2
	checkOneSided(t, rc, ex22, 2, 7, -2, 2)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValuesGap2(t *testing.T) {
	bc2, ex2 := createChannelStartEnd(2, 5, true, true)
	bc11, ex11 := createChannelStartEnd(0, 2, true, false)
	bc12, ex12 := createChannelStartEnd(5, 10, false, true)

	bc1 := appendChannels(bc11, bc12)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// sole channel 1
	checkOneSided(t, rc, ex11, 0, 2, 0, 1)

	// sole channel 2
	checkOneSided(t, rc, ex2, 2, 5, -2, 2)

	// sole channel 1
	checkOneSided(t, rc, ex12, 2, 7, -2, 1)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValuesInterleaved1(t *testing.T) {
	bc11, ex11 := createChannelStartEnd(0, 2, true, false)
	bc12, ex12 := createChannelStartEnd(4, 6, false, true)
	bc21, ex21 := createChannelStartEnd(2, 4, true, false)
	bc22, ex22 := createChannelStartEnd(6, 8, false, true)

	bc1 := appendChannels(bc11, bc12)
	bc2 := appendChannels(bc21, bc22)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// sole channel 1
	checkOneSided(t, rc, ex11, 0, 2, 0, 1)

	// sole channel 2
	checkOneSided(t, rc, ex21, 2, 4, -2, 2)

	// sole channel 1
	checkOneSided(t, rc, ex12, 4, 6, -4, 1)

	// sole channel 2
	checkOneSided(t, rc, ex22, 6, 8, -6, 2)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValuesInterleaved2(t *testing.T) {
	bc21, ex21 := createChannelStartEnd(0, 2, true, false)
	bc22, ex22 := createChannelStartEnd(4, 6, false, true)
	bc11, ex11 := createChannelStartEnd(2, 4, true, false)
	bc12, ex12 := createChannelStartEnd(6, 8, false, true)

	bc1 := appendChannels(bc11, bc12)
	bc2 := appendChannels(bc21, bc22)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// sole channel 2
	checkOneSided(t, rc, ex21, 0, 2, 0, 2)

	// sole channel 1
	checkOneSided(t, rc, ex11, 2, 4, -2, 1)

	// sole channel 2
	checkOneSided(t, rc, ex22, 4, 6, -4, 2)

	// sole channel 1
	checkOneSided(t, rc, ex12, 6, 8, -6, 1)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosOverlap1(t *testing.T) {
	bc1, ex1 := createChannel(0, 7)
	bc2, ex2 := createChannel(4, 10)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	checkOneSided(t, rc, ex1, 0, 4, 0, 1)

	checkMerged(t, rc, ex1, ex2, 4, 7, 0, -4)

	checkOneSided(t, rc, ex2, 7, 10, -4, 2)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosOverlap2(t *testing.T) {
	bc2, ex2 := createChannel(0, 7)
	bc1, ex1 := createChannel(4, 10)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	checkOneSided(t, rc, ex2, 0, 4, 0, 2)

	checkMerged(t, rc, ex1, ex2, 4, 7, -4, 0)

	checkOneSided(t, rc, ex1, 7, 10, -4, 1)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValuesInterleavedOverlap1(t *testing.T) {
	bc11, ex11 := createChannelStartEnd(0, 2, true, false)
	bc12, ex12 := createChannelStartEnd(4, 7, false, true)
	bc21, ex21 := createChannelStartEnd(2, 4, true, false)
	bc22, ex22 := createChannelStartEnd(6, 8, false, true)

	bc1 := appendChannels(bc11, bc12)
	bc2 := appendChannels(bc21, bc22)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// sole channel 1
	checkOneSided(t, rc, ex11, 0, 2, 0, 1)

	// sole channel 2
	checkOneSided(t, rc, ex21, 2, 4, -2, 2)

	// sole channel 1
	checkOneSided(t, rc, ex12, 4, 6, -4, 1)

	// merged
	checkMerged(t, rc, ex12, ex22, 6, 7, -4, -6)

	// sole channel 2
	checkOneSided(t, rc, ex22, 7, 8, -6, 2)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValuesInterleavedOverlap2(t *testing.T) {
	bc21, ex21 := createChannelStartEnd(0, 2, true, false)
	bc22, ex22 := createChannelStartEnd(4, 7, false, true)
	bc11, ex11 := createChannelStartEnd(2, 4, true, false)
	bc12, ex12 := createChannelStartEnd(6, 8, false, true)

	bc1 := appendChannels(bc11, bc12)
	bc2 := appendChannels(bc21, bc22)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// sole channel 2
	checkOneSided(t, rc, ex21, 0, 2, 0, 2)

	// sole channel 1
	checkOneSided(t, rc, ex11, 2, 4, -2, 1)

	// sole channel 2
	checkOneSided(t, rc, ex22, 4, 6, -4, 2)

	// merged
	checkMerged(t, rc, ex12, ex22, 6, 7, -6, -4)

	// sole channel 1
	checkOneSided(t, rc, ex12, 7, 8, -6, 1)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValuesInterleavedMultipleOverlap1(t *testing.T) {
	bc11, ex11 := createChannelStartEnd(0, 2, true, false)
	bc12, ex12 := createChannelStartEnd(3, 5, false, true)
	bc21, ex21 := createChannelStartEnd(0, 1, true, false)
	bc22, ex22 := createChannelStartEnd(2, 3, false, true)
	bc23, ex23 := createChannelStartEnd(5, 6, false, true)

	bc1 := appendChannels(bc11, bc12)
	bc2 := appendChannels(bc21, bc22, bc23)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// merged
	checkMerged(t, rc, ex11, ex21, 0, 1, 0, 0)

	// sole channel 1
	checkOneSided(t, rc, ex11, 1, 2, 0, 1)

	// sole channel 2
	checkOneSided(t, rc, ex22, 2, 3, -2, 2)

	// sole channel 1
	checkOneSided(t, rc, ex12, 3, 5, -3, 1)

	// sole channel 2
	checkOneSided(t, rc, ex23, 5, 6, -5, 2)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosValuesInterleavedMultipleOverlap2(t *testing.T) {
	bc21, ex21 := createChannelStartEnd(0, 2, true, false)
	bc22, ex22 := createChannelStartEnd(3, 5, false, true)
	bc11, ex11 := createChannelStartEnd(0, 1, true, false)
	bc12, ex12 := createChannelStartEnd(2, 3, false, true)
	bc13, ex13 := createChannelStartEnd(5, 6, false, true)

	bc1 := appendChannels(bc11, bc12, bc13)
	bc2 := appendChannels(bc21, bc22)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	// merged
	checkMerged(t, rc, ex11, ex21, 0, 1, 0, 0)

	// sole channel 2
	checkOneSided(t, rc, ex21, 1, 2, 0, 2)

	// sole channel 1
	checkOneSided(t, rc, ex12, 2, 3, -2, 1)

	// sole channel 2
	checkOneSided(t, rc, ex22, 3, 5, -3, 2)

	// sole channel 1
	checkOneSided(t, rc, ex13, 5, 6, -5, 1)

	checkChannelEmpty(t, rc)
}

func TestCIRatiosTwoOrdered(t *testing.T) {
	bc11, ex11 := createChannelStartEnd(0, 10, true, false)
	bc12, ex12 := createChannelStartEnd(0, 10, false, true)
	bc21, ex21 := createChannelStartEnd(0, 10, true, false)
	bc22, ex22 := createChannelStartEnd(0, 10, false, true)

	bc1 := appendChannels(bc11, bc12)
	bc2 := appendChannels(bc21, bc22)

	cif, cirf := ciFuncs(2, 1, stat.Mean, ciLevels, bench.AllInvocations)
	rc := bootstrap.CIRatios(bc1, bc2, cif, cirf)

	checkMerged(t, rc, ex11, ex21, 0, 10, 0, 0)

	checkMerged(t, rc, ex12, ex22, 10, 20, -10, -10)

	checkChannelEmpty(t, rc)
}
