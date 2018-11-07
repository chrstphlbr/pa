package main // import "bitbucket/sealuzh/pa"

import (
	"context"
	"flag"
	"fmt"
	"os"
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

func parseArgs() (c cmd, sim int, sigLev float64, statFunc statisticFunc, f1, f2 string) {
	sfStr := flag.String("st", "mean", "The statistic to be calculated")
	s := flag.Int("bs", 1000, "Number of bootstrap simulations")
	sl := flag.Float64("sig", 0.05, "Significance level")
	flag.Parse()

	args := flag.Args()
	largs := len(args)
	if largs == 1 {
		// single file -> only report confidence intervals
		c = cmdCI
		f1 = args[0]
	} else if largs == 2 {
		// two files -> report performance changes
		c = cmdDet
		f1 = args[0]
		f2 = args[1]
	} else if largs < 1 {
		fmt.Fprintf(os.Stdout, "Expected at least one file argument\n\n")
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

	return c, *s, *sl, sf, f1, f2
}

func main() {
	cmd, sim, sigLev, sf, f1, f2 := parseArgs()

	fmt.Fprintf(os.Stdout, "#Execute CIs:\n# cmd = '%s'\n# bootstrap simulations = %d\n# significance level = %.2f\n# Statistic = %s\n# file 1 = %s\n# file 2 = %s\n\n", cmd, sim, sigLev, sf.Name, f1, f2)

	nrWorkers := 1000

	var exec func()
	switch cmd {
	case cmdCI:
		exec = func() {
			ci(sim, nrWorkers, sigLev, sf.Func, f1)
		}
	case cmdDet:
		exec = func() {
			det(sim, nrWorkers, sigLev, sf.Func, f1, f2)
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

func ci(sim int, nrWorkers int, sigLev float64, sf stat.StatisticFunc, fp string) {
	f, err := os.Open(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open file '%s'", fp)
	}

	c, err := bench.FromCSV(context.Background(), f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	rc := bootstrap.CIs(c, sim, nrWorkers, sf, sigLev)

	for res := range rc {
		if res.Err != nil {
			fmt.Fprintf(os.Stderr, "Error while retrieving CI result: %v", res.Err)
		}

		b := res.Benchmark
		ci := res.CI
		fmt.Fprintf(os.Stdout, "%s;%s;%s;%e;%e;%.2f\n", b.Name, b.FunctionParams, b.PerfParams, ci.Lower, ci.Upper, ci.Level)
	}
}

func det(sim int, nrWorkers int, sigLev float64, sf stat.StatisticFunc, fp1, fp2 string) {
	panic("Not implemented")
}
