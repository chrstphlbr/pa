package bootstrap

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/chrstphlbr/pa/pkg/bench"
	st "github.com/chrstphlbr/pa/pkg/stat"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
	"gonum.org/v1/gonum/stat/sampleuv"
)

type CIFunc = func(bench.ExecutionSlice) []st.CI
type CIRatioFunc = func(bench.ExecutionSlice, bench.ExecutionSlice) []st.CIRatio

func CIRatioFuncSetup(iters int, maxNrWorkers int, statFunc st.StatisticFunc, significanceLevels []float64, sampler bench.InvocationSampler) CIRatioFunc {
	return func(executionsA bench.ExecutionSlice, executionsB bench.ExecutionSlice) []st.CIRatio {
		return CIRatio(iters, maxNrWorkers, statFunc, significanceLevels, executionsA, executionsB, sampler)
	}
}

func CIFuncSetup(iters int, maxNrWorkers int, statFunc st.StatisticFunc, significanceLevels []float64, sampler bench.InvocationSampler) CIFunc {
	return func(executions bench.ExecutionSlice) []st.CI {
		return CI(iters, maxNrWorkers, statFunc, significanceLevels, executions, sampler)
	}
}

func CIRatio(iters int, maxNrWorkers int, statisticFunc st.StatisticFunc, significanceLevels []float64, executionsA bench.ExecutionSlice, executionsB bench.ExecutionSlice, sampler bench.InvocationSampler) []st.CIRatio {
	metricA, simStatA := metricAndSimulations(iters, maxNrWorkers, statisticFunc, executionsA, sampler)
	ciAs := ci(metricA, simStatA, significanceLevels)

	metricB, simStatB := metricAndSimulations(iters, maxNrWorkers, statisticFunc, executionsB, sampler)
	ciBs := ci(metricB, simStatB, significanceLevels)

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
	ratioMetric := statisticFunc(ratios)
	ciRatios := ci(ratioMetric, ratios, significanceLevels)

	lsl := len(significanceLevels)
	ret := make([]st.CIRatio, lsl)
	for i := 0; i < lsl; i++ {
		ret[i] = st.CIRatio{
			CIA:     ciAs[i],
			CIB:     ciBs[i],
			CIRatio: ciRatios[i],
		}
	}
	return ret
}

func CI(iters int, maxNrWorkers int, statisticFunc st.StatisticFunc, significanceLevels []float64, executions bench.ExecutionSlice, sampler bench.InvocationSampler) []st.CI {
	metric, simStat := metricAndSimulations(iters, maxNrWorkers, statisticFunc, executions, sampler)
	return ci(metric, simStat, significanceLevels)
}

func metricAndSimulations(iters int, maxNrWorkers int, statisticFunc st.StatisticFunc, executions bench.ExecutionSlice, sampler bench.InvocationSampler) (metric float64, simStat []float64) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		metric = benchMetric(executions, statisticFunc)
		wg.Done()
	}()
	go func() {
		simStat = simulatedStatistics(iters, maxNrWorkers, statisticFunc, executions, sampler)
		wg.Done()
	}()

	wg.Wait()

	return metric, simStat
}

func benchMetric(executions bench.ExecutionSlice, statisticFunc st.StatisticFunc) float64 {
	meanIterations := executions.FlatSlice(bench.MeanInvocations)
	metric := statisticFunc(meanIterations)
	return metric
}

func ci(metric float64, d []float64, significanceLevels []float64) []st.CI {
	sort.Float64s(d)
	lstat := float64(len(d))

	ret := make([]st.CI, len(significanceLevels))
	for i, significanceLevel := range significanceLevels {

		sl := st.SigLevel(significanceLevel)

		slhalf := sl / 2
		clhalf := 1 - slhalf

		lqi := int(math.Ceil(lstat * slhalf))
		uqi := int(math.Floor(lstat * clhalf))

		lq := d[lqi]
		uq := d[uqi]

		ret[i] = st.CI{
			Metric: metric,
			Lower:  lq,
			Upper:  uq,
			Level:  1 - sl,
		}
	}
	return ret
}

func simulatedStatistics(iters int, maxNrWorkers int, statisticFunc st.StatisticFunc, executions bench.ExecutionSlice, sampler bench.InvocationSampler) []float64 {
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
						rs := randomResampling(executions, sampler)
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

func randomResampling(d bench.ExecutionSlice, sampler bench.InvocationSampler) []float64 {
	s := d.Slice(sampler)

	lis := len(s)
	isample := sampleSize(lis)

	var ret []float64
	// ret := make([]float64, 0, d.ElementCount())

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
