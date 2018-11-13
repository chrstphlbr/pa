package main // import "bitbucket/sealuzh/pa"

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"bitbucket.org/sealuzh/pa/pkg/bootstrap"

	"bitbucket.org/sealuzh/pa/pkg/bench"

	"bitbucket.org/sealuzh/pa/pkg/stat"
)

type cmd int

func (c cmd) String() string {
	switch c {
	case 0:
		return "ci"
	case 1:
		return "det"
	}
	return "INVALID_COMMAND"
}

const (
	cmdCI cmd = iota
	cmdDet
)

type statisticFunc struct {
	Name string
	Func stat.StatisticFunc
}

func parseArgs() (c cmd, sim int, sigLev float64, statFunc statisticFunc, f1, f2 []string, sampler bench.InvocationSampler, printMem bool) {
	sfStr := flag.String("st", "mean", "The statistic to be calculated")
	s := flag.Int("bs", 1000, "Number of bootstrap simulations")
	sl := flag.Float64("sig", 0.05, "Significance level")
	is := flag.Int("is", -1, "Number of invocation samples taken (-1 for all)")
	sm := flag.Bool("sm", false, "Take mean of all invocation samples. If this flag is provided, -is is invalid")
	m := flag.Int("m", 1, "Number of multiple files belongig to one group (test or control); e.g., 3 means 6 files in total, 3 test and 3 control")
	rm := flag.Bool("mem", false, "Print runtime memory to Stdout")
	flag.Parse()

	args := flag.Args()
	largs := len(args)
	if largs == 1 {
		// single file -> only report confidence intervals
		c = cmdCI
		f1 = []string{args[0]}
	} else if largs == 2 {
		// two files -> report performance changes
		c = cmdDet
		f1 = []string{args[0]}
		f2 = []string{args[1]}
	} else if largs < 1 {
		fmt.Fprintf(os.Stdout, "Expected at least one file argument\n\n")
		flag.Usage()
		os.Exit(1)
	} else {
		c = cmdDet
		// multiple files for test and control group
		if largs / *m != 2 {
			fmt.Fprintf(os.Stdout, "-m must be half of number of arguments\n\n")
			flag.Usage()
			os.Exit(1)
		}

		f1 = []string{}
		for i := 0; i < *m; i++ {
			f1 = append(f1, args[i])
		}
		f2 = []string{}
		for i := *m; i < *m*2; i++ {
			f2 = append(f2, args[i])
		}
	}

	if *is < -1 || *is == 0 {
		fmt.Fprint(os.Stdout, "Invalid number of invocation samples, must be > 0 (for numer of samples) or -1 for all\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *sm {
		sampler = bench.MeanInvocations
	} else if *is == -1 {
		sampler = bench.AllInvocations
	} else {
		sampler = bench.SampleInvocations(*is)
	}

	statisticFunction := *sfStr
	var sf statisticFunc
	switch statisticFunction {
	case "mean":
		sf = statisticFunc{
			Name: "Mean",
			Func: stat.Mean,
		}
	case "cov":
		sf = statisticFunc{
			Name: "COV",
			Func: stat.COV,
		}
	case "median":
		sf = statisticFunc{
			Name: "Median",
			Func: stat.Median,
		}
	default:
		fmt.Fprintf(os.Stdout, "Unknown statistics function '%s'\n\n", statisticFunction)
		flag.Usage()
		os.Exit(1)
	}

	return c, *s, *sl, sf, f1, f2, sampler, *rm
}

func main() {
	cmd, sim, sigLev, sf, f1, f2, sampler, printMem := parseArgs()
	maxNrWorkers := runtime.NumCPU()

	fmt.Fprintf(os.Stdout, "#Execute CIs:\n# cmd = '%s'\n# number of cores = %d\n# bootstrap simulations = %d\n# significance level = %.2f\n# Statistic = %s\n# file 1 = %s\n# file 2 = %s\n\n", cmd, maxNrWorkers, sim, sigLev, sf.Name, f1, f2)

	ciFunc := bootstrap.CIFuncSetup(sim, maxNrWorkers, sf.Func, sigLev, sampler)
	ciRatioFunc := bootstrap.CIRatioFuncSetup(sim, maxNrWorkers, sf.Func, sigLev, sampler)

	var exec func()
	switch cmd {
	case cmdCI:
		exec = func() {
			ci(ciFunc, f1[0], printMem)
		}
	case cmdDet:
		exec = func() {
			det(ciFunc, ciRatioFunc, f1, f2, printMem)
		}
	default:
		fmt.Fprintf(os.Stdout, "Invalid command '%s' (available: 'ci' and 'det')\n\n", cmd)
		flag.Usage()
		os.Exit(1)
	}

	start := time.Now()
	exec()
	fmt.Fprintf(os.Stdout, "#Total execution took %v\n", time.Since(start))
}

func ci(ciFunc bootstrap.CIFunc, fp string, printMem bool) {
	f, err := os.Open(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open file '%s'", fp)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := bench.FromCSV(ctx, f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return
	}

	rc := bootstrap.CIs(c, ciFunc)

	printMemStats(printMem)

	for res := range rc {
		if res.Err != nil {
			fmt.Fprintf(os.Stderr, "Error while retrieving CI result: %v", res.Err)
		}

		b := res.Benchmark
		ci := res.CI
		fmt.Fprintf(os.Stdout, "%s;%s;%s;%e;%e;%.2f\n", b.Name, b.FunctionParams, b.PerfParams, ci.Lower, ci.Upper, ci.Level)
		printMemStats(printMem)
	}
}

func det(ciFunc bootstrap.CIFunc, ciRatioFunc bootstrap.CIRatioFunc, fp1, fp2 []string, printMem bool) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c1, err := mergedInput(ctx, fp1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	c2, err := mergedInput(ctx, fp2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	rc := bootstrap.CIRatios(c1, c2, ciFunc, ciRatioFunc)

	printMemStats(printMem)

	for res := range rc {
		if res.Err != nil {
			fmt.Fprintf(os.Stderr, "Error while retrieving CI result: %v\n", res.Err)
			return
		}

		b := res.Benchmark
		cir := res.CIRatio
		fmt.Fprintf(
			os.Stdout,
			"%s;%s;%s;%e;%e;%.2f;%e;%e;%.2f;%e;%e;%.2f\n",
			b.Name, b.FunctionParams, b.PerfParams,
			cir.CIA.Lower, cir.CIA.Upper, cir.CIA.Level,
			cir.CIB.Lower, cir.CIB.Upper, cir.CIB.Level,
			cir.CIRatio.Lower, cir.CIRatio.Upper, cir.CIRatio.Level,
		)
		printMemStats(printMem)
	}
}

func mergedInput(ctx context.Context, fs []string) (bench.Chan, error) {
	var chans []bench.Chan
	for _, fn := range fs {
		f, err := os.Open(fn)
		if err != nil {
			return nil, fmt.Errorf("Could not open file1 '%s'", fn)
		}

		c1, err := bench.FromCSV(ctx, f)
		if err != nil {
			return nil, fmt.Errorf("Could not read from CSV for file '%s': %v", fn, err)
		}
		chans = append(chans, c1)
	}
	return bench.MergeChans(chans...), nil
}

func printMemStats(print bool) {
	if print {
		ms := &runtime.MemStats{}
		runtime.ReadMemStats(ms)
		fmt.Fprintf(os.Stdout, "# current memory consumption: sys=%d, heapAlloc=%d, heapInuse=%d, stackInuse=%d, numGCs=%d\n", ms.Sys, ms.HeapAlloc, ms.HeapInuse, ms.StackInuse, ms.NumGC)
	}
}
