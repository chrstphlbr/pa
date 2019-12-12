package main // import "github.com/chrstphlbr/pa"

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/chrstphlbr/pa/pkg/bootstrap"

	"github.com/chrstphlbr/pa/pkg/bench"

	"github.com/chrstphlbr/pa/pkg/stat"
)

type cmd int

func (c cmd) String() string {
	switch c {
	case 0:
		return "CI"
	case 1:
		return "Detection"
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

func parseArgs() (c cmd, sim int, sigLevs []float64, statFunc statisticFunc, f1, f2 []string, invocationSamples int, outputMetric bool, printMem bool) {
	sfStr := flag.String("st", "mean", "The statistic to be calculated")
	s := flag.Int("bs", 10000, "Number of bootstrap simulations")
	sls := flag.String("sl", "0.01", "Significance levels (multiple seperated by ',')")
	is := flag.Int("is", 0, "Number of invocation samples (0 for mean across all invocations, -1 for all, > 0 for number of samples)")
	m := flag.Int("m", 1, "Number of multiple files belongig to one group (test or control); e.g., 3 means 6 files in total, 3 test and 3 control")
	om := flag.Bool("os", false, "Include statistic (e.g., mean) in output")
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

	// significance levels
	slsSplitted := strings.Split(*sls, ",")
	slsFloat := make([]float64, 0, len(slsSplitted))
	for _, sl := range slsSplitted {
		slFloat, err := strconv.ParseFloat(sl, 64)
		if err != nil {
			fmt.Fprintf(os.Stdout, "Could not parse significance level '%s' into float64: %v\n\n", sl, err)
			flag.Usage()
			os.Exit(1)
		}
		slsFloat = append(slsFloat, slFloat)
	}

	if *is < -1 {
		fmt.Fprint(os.Stdout, "Invalid number of invocation samples, must be 0 for mean across samples from iteration, > 0 for number of samples, or -1 for all\n\n")
		flag.Usage()
		os.Exit(1)
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

	return c, *s, slsFloat, sf, f1, f2, *is, *om, *rm
}

func main() {
	cmd, sim, sigLevels, sf, f1, f2, is, outputMetric, printMem := parseArgs()
	maxNrWorkers := runtime.NumCPU()

	var sampler bench.InvocationSampler
	var samplingType string
	if is == 0 {
		sampler = bench.MeanInvocations
		samplingType = "Mean"
	} else if is == -1 {
		sampler = bench.AllInvocations
		samplingType = "All"
	} else {
		sampler = bench.SampleInvocations(is)
		samplingType = fmt.Sprintf("%d invocations per iteration", is)
	}

	var outHeader strings.Builder
	outHeader.WriteString("#Execute CIs:\n")
	outHeader.WriteString(fmt.Sprintf("# cmd = %s\n", cmd))
	outHeader.WriteString(fmt.Sprintf("# number of cores = %d\n", maxNrWorkers))
	outHeader.WriteString(fmt.Sprintf("# bootstrap simulations = %d\n", sim))
	outHeader.WriteString(fmt.Sprintf("# significance levels = %v\n", sigLevels))
	outHeader.WriteString(fmt.Sprintf("# statistic = %s\n", sf.Name))
	outHeader.WriteString(fmt.Sprintf("# include statistic in output = %t\n", outputMetric))
	outHeader.WriteString(fmt.Sprintf("# invocation sampling = %s\n", samplingType))
	outHeader.WriteString(fmt.Sprintf("# files 1 = %s\n", f1))
	outHeader.WriteString(fmt.Sprintf("# files 2 = %s\n", f2))
	fmt.Fprint(os.Stdout, outHeader.String())
	fmt.Fprintln(os.Stdout, "")

	ciFunc := bootstrap.CIFuncSetup(sim, maxNrWorkers, sf.Func, sigLevels, sampler)
	ciRatioFunc := bootstrap.CIRatioFuncSetup(sim, maxNrWorkers, sf.Func, sigLevels, sampler)

	var exec func()
	switch cmd {
	case cmdCI:
		exec = func() {
			ci(ciFunc, f1[0], outputMetric, printMem)
		}
	case cmdDet:
		exec = func() {
			det(ciFunc, ciRatioFunc, f1, f2, outputMetric, printMem)
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

func ci(ciFunc bootstrap.CIFunc, fp string, outputMetric, printMem bool) {
	f, err := os.Open(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open file '%s'\n", fp)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := bench.FromCSV(ctx, f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	rc := bootstrap.CIs(c, ciFunc)

	printMemStats(printMem)

	for res := range rc {
		if res.Err != nil {
			fmt.Fprintf(os.Stderr, "Error while retrieving CI result: %v", res.Err)
			continue
		}

		b := res.Benchmark
		cis := res.CIs
		for _, ci := range cis {
			if outputMetric {
				// include statistic/metric in output
				fmt.Fprintf(os.Stdout, "%s;%s;%s;%e;%e;%e;%.2f\n", b.Name, b.FunctionParams, b.PerfParams, ci.Metric, ci.Lower, ci.Upper, ci.Level)
			} else {
				// only print CIs
				fmt.Fprintf(os.Stdout, "%s;%s;%s;%e;%e;%.2f\n", b.Name, b.FunctionParams, b.PerfParams, ci.Lower, ci.Upper, ci.Level)
			}
		}
		printMemStats(printMem)
	}
}

func det(ciFunc bootstrap.CIFunc, ciRatioFunc bootstrap.CIRatioFunc, fp1, fp2 []string, outputMetric, printMem bool) {
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
			continue
		}

		b := res.Benchmark
		cirs := res.CIRatios
		for _, cir := range cirs {
			if outputMetric {
				// include statistic/metric in output
				fmt.Fprintf(
					os.Stdout,
					"%s;%s;%s;%e;%e;%e;%.2f;%e;%e;%e;%.2f;%e;%e;%e;%.2f\n",
					b.Name, b.FunctionParams, b.PerfParams,
					cir.CIA.Metric, cir.CIA.Lower, cir.CIA.Upper, cir.CIA.Level,
					cir.CIB.Metric, cir.CIB.Lower, cir.CIB.Upper, cir.CIB.Level,
					cir.CIRatio.Metric, cir.CIRatio.Lower, cir.CIRatio.Upper, cir.CIRatio.Level,
				)
			} else {
				// only print CIs
				fmt.Fprintf(
					os.Stdout,
					"%s;%s;%s;%e;%e;%.2f;%e;%e;%.2f;%e;%e;%.2f\n",
					b.Name, b.FunctionParams, b.PerfParams,
					cir.CIA.Lower, cir.CIA.Upper, cir.CIA.Level,
					cir.CIB.Lower, cir.CIB.Upper, cir.CIB.Level,
					cir.CIRatio.Lower, cir.CIRatio.Upper, cir.CIRatio.Level,
				)
			}
		}
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
