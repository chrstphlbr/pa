package bootstrap

import (
	"fmt"

	"bitbucket.org/sealuzh/pa/pkg/bench"
	"bitbucket.org/sealuzh/pa/pkg/stat"
)

type CIResult struct {
	Benchmark *bench.B
	CI        stat.CI
	Err       error
}

func CIs(c bench.Chan, iters int, maxNrWorkers int, statFunc stat.StatisticFunc, significanceLevel float64) <-chan CIResult {
	out := make(chan CIResult)
	go func() {
		defer close(out)
		for br := range c {
			switch br.Type {
			case bench.ExecError:
				out <- CIResult{
					Err: br.Err,
				}
			case bench.ExecNext:
				out <- CIResult{
					Benchmark: br.Exec.Benchmark,
					CI:        CI(iters, maxNrWorkers, statFunc, br.Exec, significanceLevel),
				}
			}
		}
	}()
	return out
}

type CIRatioResult struct {
	Benchmark *bench.B
	CIRatio   stat.CIRatio
	Err       error
}

type chanNumber int

const (
	cNr1 chanNumber = iota
	cNr2
)

type leftOver struct {
	ev  *bench.ExecutionValue
	cnr chanNumber
}

func CIRatios(c1, c2 bench.Chan, iters int, maxNrWorkers int, statFunc stat.StatisticFunc, significanceLevel float64) <-chan CIRatioResult {
	out := make(chan CIRatioResult)

	go func() {
		defer close(out)

		var leftOver *leftOver
		for {
			var ev1, ev2 *bench.ExecutionValue
			var ok1, ok2 bool

			if leftOver != nil {
				// has leftOver -> only read from one channel
				if leftOver.cnr == cNr1 {
					ev1 = leftOver.ev
					ok1 = true
					ev2v, ok2v := <-c2
					ev2 = &ev2v
					ok2 = ok2v
					// fmt.Fprintf(os.Stderr, "leftOver1: %v  %t\n\t%v  %t\n", ev1, ok1, ev2, ok2)
				} else if leftOver.cnr == cNr2 {
					ev2 = leftOver.ev
					ok2 = true
					ev1v, ok1v := <-c1
					ev1 = &ev1v
					ok1 = ok1v
					// fmt.Fprintf(os.Stderr, "leftOver2: %v  %t\n\t%v  %t\n", ev1, ok1, ev2, ok2)
				} else {
					panic(fmt.Sprintf("Invalid channel number %d", leftOver.cnr))
				}
				// needed if files are not ordered
				leftOver = nil
			} else {
				// no leftOver -> read from both channels
				ev1v, ok1v := <-c1
				ev1 = &ev1v
				ok1 = ok1v
				ev2v, ok2v := <-c2
				ev2 = &ev2v
				ok2 = ok2v
			}

			if ok1 && ok2 {
				// both values received
				leftOver = handleTwoResults(out, ev1, ev2, iters, maxNrWorkers, statFunc, significanceLevel)
			} else if ok1 {
				// only c1 received
				handleSingleResult(out, ev1, cNr1, iters, maxNrWorkers, statFunc, significanceLevel)
			} else if ok2 {
				// only c2 received
				handleSingleResult(out, ev2, cNr2, iters, maxNrWorkers, statFunc, significanceLevel)
			} else {
				// done
				break
			}
		}
	}()

	return out
}

func handleSingleResult(out chan<- CIRatioResult, ev *bench.ExecutionValue, cnr chanNumber, iters int, maxNrWorkers int, statFunc stat.StatisticFunc, significanceLevel float64) {
	if ev.Type == bench.ExecStart || ev.Type == bench.ExecEnd {
		return
	} else if ev.Type == bench.ExecError {
		out <- CIRatioResult{
			Err: ev.Err,
		}
		return
	}

	ci := CI(iters, maxNrWorkers, statFunc, ev.Exec, significanceLevel)

	cir := stat.CIRatio{}
	if cnr == cNr1 {
		cir.CIA = ci
	} else if cnr == cNr2 {
		cir.CIB = ci
	} else {
		// invalid channel number cnr
		panic(fmt.Sprintf("Invalid channel number %d", cnr))
	}

	out <- CIRatioResult{
		Benchmark: ev.Exec.Benchmark,
		CIRatio:   cir,
	}
}

func handleTwoResults(out chan<- CIRatioResult, ev1, ev2 *bench.ExecutionValue, iters int, maxNrWorkers int, statFunc stat.StatisticFunc, significanceLevel float64) *leftOver {
	if (ev1.Type == bench.ExecStart || ev1.Type == bench.ExecEnd) && (ev2.Type == bench.ExecStart || ev2.Type == bench.ExecEnd) {
		// handle both started or both done
		return nil
	}

	if ev1.Type == bench.ExecError {
		// channel 1 sent error
		out <- CIRatioResult{
			Err: ev1.Err,
		}
		handleSingleResult(out, ev2, cNr2, iters, maxNrWorkers, statFunc, significanceLevel)
	} else if ev2.Type == bench.ExecError {
		// channel 1 sent error
		out <- CIRatioResult{
			Err: ev2.Err,
		}
		handleSingleResult(out, ev1, cNr1, iters, maxNrWorkers, statFunc, significanceLevel)
	} else if ev1.Type == bench.ExecEnd {
		// channel 1 is done
		handleSingleResult(out, ev2, cNr2, iters, maxNrWorkers, statFunc, significanceLevel)
	} else if ev2.Type == bench.ExecEnd {
		// channel 2 is done
		handleSingleResult(out, ev1, cNr1, iters, maxNrWorkers, statFunc, significanceLevel)
	} else {
		// both channels have a result
		return handleTwoValidResults(out, ev1, ev2, iters, maxNrWorkers, statFunc, significanceLevel)
	}

	return nil
}

func handleTwoValidResults(out chan<- CIRatioResult, ev1, ev2 *bench.ExecutionValue, iters int, maxNrWorkers int, statFunc stat.StatisticFunc, significanceLevel float64) *leftOver {
	ex1 := ev1.Exec
	ex2 := ev2.Exec

	cmp := ex1.Benchmark.Compare(ex2.Benchmark)

	switch cmp {
	case 0:
		out <- CIRatioResult{
			Benchmark: ex1.Benchmark,
			CIRatio:   CIRatio(iters, maxNrWorkers, statFunc, ex1, ex2, significanceLevel),
		}
	case -1:
		handleSingleResult(out, ev1, cNr1, iters, maxNrWorkers, statFunc, significanceLevel)
		return &leftOver{
			ev:  ev2,
			cnr: cNr2,
		}
	case 1:
		handleSingleResult(out, ev2, cNr2, iters, maxNrWorkers, statFunc, significanceLevel)
		return &leftOver{
			ev:  ev1,
			cnr: cNr1,
		}
	default:
		panic(fmt.Sprintf("Invalid result from Compare: %d", cmp))
	}
	return nil
}
