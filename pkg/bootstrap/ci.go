package bootstrap

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"bitbucket.org/sealuzh/pa/pkg/bench"
	st "bitbucket.org/sealuzh/pa/pkg/stat"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
	"gonum.org/v1/gonum/stat/sampleuv"
)

func CIRatioFunc(iters int, maxNrWorkers int) st.CIRatioFunc {
	return func(executionsA bench.ExecutionSlice, executionsB bench.ExecutionSlice, statFunc st.StatisticFunc, significanceLevel float64) st.CIRatio {
		return CIRatio(iters, maxNrWorkers, statFunc, executionsA, executionsB, significanceLevel)
	}
}

func CIFunc(iters int, maxNrWorkers int) st.CIFunc {
	return func(executions bench.ExecutionSlice, statFunc st.StatisticFunc, significanceLevel float64) st.CI {
		return CI(iters, maxNrWorkers, statFunc, executions, significanceLevel)
	}
}

func CIRatio(iters int, maxNrWorkers int, statisticFunc st.StatisticFunc, executionsA bench.ExecutionSlice, executionsB bench.ExecutionSlice, significanceLevel float64) st.CIRatio {
	simStatA := simulatedStatistics(iters, maxNrWorkers, statisticFunc, executionsA)
	ciA := ci(simStatA, significanceLevel)

	simStatB := simulatedStatistics(iters, maxNrWorkers, statisticFunc, executionsB)
	ciB := ci(simStatB, significanceLevel)

	lSimA := len(simStatA)
	lSimB := len(simStatB)
	if lSimA != lSimB {
		panic(fmt.Sprintf("Simulated statistics not of same size: len(a) = %d; len(b) = %d", lSimA, lSimB))
	}

	ratios := make([]float64, 0, lSimA)
	for i := 0; i < iters; i++ {
		ratio := simStatB[i] / simStatA[i]
		ratios = append(ratios, ratio)
	}

	return st.CIRatio{
		CIA:     ciA,
		CIB:     ciB,
		CIRatio: ci(ratios, significanceLevel),
	}
}

func CI(iters int, maxNrWorkers int, statisticFunc st.StatisticFunc, executions bench.ExecutionSlice, significanceLevel float64) st.CI {
	simStat := simulatedStatistics(iters, maxNrWorkers, statisticFunc, executions)
	return ci(simStat, significanceLevel)
}

func ci(d []float64, significanceLevel float64) st.CI {
	sl := st.SigLevel(significanceLevel)

	slhalf := sl / 2
	clhalf := 1 - sl

	sort.Float64s(d)
	lstat := float64(len(d))

	lqi := int(math.Ceil(lstat * slhalf))
	uqi := int(math.Floor(lstat * clhalf))

	lq := d[lqi]
	uq := d[uqi]

	return st.CI{
		Lower: lq,
		Upper: uq,
		Level: 1 - sl,
	}
}

func simulatedStatistics(iters int, maxNrWorkers int, statisticFunc st.StatisticFunc, executions bench.ExecutionSlice) []float64 {
	// create workers
	var wg sync.WaitGroup
	wg.Add(iters)

	var anw int
	if iters < maxNrWorkers {
		anw = iters
	} else {
		anw = maxNrWorkers
	}

	workChan := make(chan int, iters)
	samplingChan := make(chan []float64, iters)
	for i := 0; i < anw; i++ {
		go func() {
		Loop:
			for {
				select {
				case _, ok := <-workChan:
					if ok {
						rs := randomResampling(executions)
						samplingChan <- rs
						wg.Done()
					} else {
						break Loop
					}
				}
			}
		}()
	}

	// handle simulations
	for i := 0; i < iters; i++ {
		workChan <- i
	}
	close(workChan)

	wg.Wait()
	close(samplingChan)

	// receive from sampling channel
	simStat := make([]float64, 0, iters)
	for randomSample := range samplingChan {
		stat := statisticFunc(randomSample)
		simStat = append(simStat, stat)
	}

	return simStat
}

func randomResampling(d bench.ExecutionSlice) []float64 {
	s := d.Slice()

	lis := len(s)
	isample := sampleSize(lis)

	ret := make([]float64, 0, d.ElementCount())

	for _, i := range isample {
		trials := s[i]
		ltrials := len(trials)
		tsample := sampleSize(ltrials)

		for _, t := range tsample {
			forks := trials[t]
			lforks := len(forks)
			fsample := sampleSize(lforks)

			for _, f := range fsample {
				iterations := forks[f]
				literations := len(iterations)
				itsample := sampleSize(literations)

				for _, it := range itsample {
					invocations := iterations[it]
					linvocations := len(invocations)
					invsample := sampleSize(linvocations)

					for _, inv := range invsample {
						ret = append(ret, invocations[inv])
					}
				}
			}
		}
	}

	return ret
}

func sample(d []float64) []float64 {
	nd := distuv.Normal{
		Mu:    stat.Mean(d, nil),
		Sigma: stat.StdDev(d, nil),
		Src:   rand.NewSource(uint64(time.Now().UnixNano())),
	}

	sampler := sampleuv.IIDer{
		Dist: nd,
	}

	sampler.Sample(d)
	return d
}

func sampleSize(l int) []int {
	if l == 0 {
		return []int{}
	} else if l == 1 {
		return []int{0}
	}

	id := make([]float64, 0, l)
	for dp := 0; dp < l; dp++ {
		id = append(id, float64(dp))
	}

	sampled := sample(id)

	ret := make([]int, 0, l)
	lf64 := float64(l)
	for _, dp := range sampled {
		ret = append(ret, int(math.Mod(math.Abs(dp), lf64)))
	}
	return ret
}
