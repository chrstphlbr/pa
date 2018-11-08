package bootstrap

import (
	"bitbucket.org/sealuzh/pa/pkg/bench"
	"bitbucket.org/sealuzh/pa/pkg/stat"
)

type CIResult struct {
	Benchmark *bench.B
	CI        stat.CI
	Err       error
}

func CIs(c bench.Chan, iters int, nrWorkers int, statFunc stat.StatisticFunc, significanceLevel float64) <-chan CIResult {
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
					CI:        CI(iters, nrWorkers, statFunc, br.Exec, significanceLevel),
				}
			}
		}
	}()
	return out
}
