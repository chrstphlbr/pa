package bootstrap

import (
	"bitbucket.org/sealuzh/pa/pkg/bench"
	"bitbucket.org/sealuzh/pa/pkg/stat"
)

type CIResult struct {
	Benchmark bench.B
	CI        stat.CI
	Err       error
}

func CIs(c bench.Chan, iters int, nrWorkers int, statFunc stat.StatisticFunc, significanceLevel float64) <-chan CIResult {
	res := make(chan CIResult)
	go func() {
		for br := range c {
			switch br.Type {
			case bench.ExecError:
				res <- CIResult{
					Err: br.Err,
				}
			case bench.ExecNext:
				res <- CIResult{
					Benchmark: br.Exec.Benchmark,
					CI:        CI(iters, nrWorkers, statFunc, br.Exec, significanceLevel),
				}
			}
		}
		close(res)
	}()
	return res
}
