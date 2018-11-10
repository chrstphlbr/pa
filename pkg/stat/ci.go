package stat

import (
	"fmt"
	"math"
	"sort"

	"bitbucket.org/sealuzh/pa/pkg/bench"
	"gonum.org/v1/gonum/stat"
)

type StatisticFunc func([]float64) float64
type CIFunc = func(bench.ExecutionSlice, StatisticFunc, float64) CI
type CIRatioFunc = func(bench.ExecutionSlice, bench.ExecutionSlice, StatisticFunc, float64) CIRatio

type CI struct {
	Lower float64
	Upper float64
	Level float64
}

type CIRatio struct {
	CIRatio CI
	CIA     CI
	CIB     CI
}

func SigLevel(l float64) float64 {
	var sl float64
	if l >= 0 && l <= 1 {
		sl = l
	} else if l < 0 {
		sl = 0
	} else if l > 1 {
		sl = 1
	} else {
		panic(fmt.Sprintf("Invalid significance level %f should never happen", l))
	}
	return sl
}

func Mean(data []float64) float64 {
	return stat.Mean(data, nil)
}

func Median(data []float64) float64 {
	l := len(data)
	if l == 0 {
		return 0
	} else if l == 1 {
		return data[0]
	}

	cp := make([]float64, l)
	copy(cp, data)
	sort.Float64s(cp)

	middle := (float64(l) + 1) / 2
	lower := math.Floor(middle) - 1
	upper := math.Ceil(middle) - 1

	if lower == upper {
		return cp[int(lower)]
	}

	le := cp[int(lower)]
	ue := cp[int(upper)]
	return (le + ue) / 2
}

func COV(data []float64) float64 {
	sd := StdDev(data)
	m := Mean(data)
	return sd / m
}

func StdDev(data []float64) float64 {
	return stat.StdDev(data, nil)
}
