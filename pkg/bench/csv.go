package bench

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func FromCSV(ctx context.Context, r io.Reader) (Chan, error) {
	cr := csv.NewReader(r)
	if cr == nil {
		return nil, fmt.Errorf("Could not create reader")
	}

	cr.Comma = ';'
	cr.FieldsPerRecord = 12
	cr.ReuseRecord = true

	c := make(Chan)

	// remove header
	_, err := cr.Read()
	if err != nil {
		if err == io.EOF {
			go func() {
				defer close(c)
				c <- ExecutionValue{Type: ExecStart}
				c <- ExecutionValue{Type: ExecEnd}
			}()
			return c, nil
		}
		return nil, err
	}

	go func() {
		defer close(c)

		var first *InvocationsFlat
		ev := ExecutionValue{Type: ExecStart}
		var cnt int
	Loop:
		for {
			select {
			case c <- ev:
				ev, first = parseExecution(cr, first)
				if ev.Type == ExecEnd {
					break Loop
				}
			case <-ctx.Done():
				break Loop
			}
			cnt++
		}
	}()

	return c, nil
}

func parseExecution(cr *csv.Reader, first *InvocationsFlat) (ExecutionValue, *InvocationsFlat) {
	res := ExecutionValue{
		Type: ExecNext,
	}

	// add first invocations from current benchmark
	if first != nil {
		res.Exec = NewExecution(first.Benchmark)
		err := res.Exec.AddInvocations(*first)
		if err != nil {
			return ExecutionValue{
				Type: ExecError,
				Err:  err,
			}, nil
		}
	}

	for {
		rec, err := cr.Read()

		if err != nil {
			// handle end of file
			if err == io.EOF {
				// handle EOF if there is a last element
				if res.Exec != nil {
					return res, nil
				}
				return ExecutionValue{Type: ExecEnd}, nil
			}
			// send error over channel
			return ExecutionValue{
				Type: ExecError,
				Err:  err,
			}, nil
		}

		cr, err := csvBenchExec(rec)
		if err != nil {
			// send error over channel
			return ExecutionValue{
				Type: ExecError,
				Err:  err,
			}, nil
		}

		// check benchmark; handle first benchmark of all benchmarks in result
		if res.Exec == nil {
			// first line
			res.Exec = NewExecution(cr.Benchmark)
		}

		// new benchmark
		if !cr.Benchmark.Equals(res.Exec.Benchmark) {
			return res, cr
		}

		// still same benchmark -> append to existing results
		err = res.Exec.AddInvocations(*cr)
		if err != nil {
			return ExecutionValue{
				Type: ExecError,
				Err:  err,
			}, nil
		}
	}
}

func csvBenchExec(rec []string) (*InvocationsFlat, error) {
	b := New(rec[2])
	b.FunctionParams = make(FunctionParams, 0)

	// params
	if rawps := rec[3]; rawps != "" {
		rawpsSplitted := strings.Split(rawps, ",")
		for _, rawp := range rawpsSplitted {
			p := strings.Split(rawp, "=")
			if len(p) != 2 {
				return nil, fmt.Errorf("Invalid JMH parameter (%s): '%s'", b.Name, rawp)
			}
			b.PerfParams.Add(p[0], p[1])
		}
	}

	// project;commit;benchmark;params;instance;trial;fork;iteration;mode;unit;value_count;value

	// trial
	t, err := strconv.Atoi(rec[5])
	if err != nil {
		return nil, fmt.Errorf("Could not parse 'trial': %v", err)
	}

	// fork
	f, err := strconv.Atoi(rec[6])
	if err != nil {
		return nil, fmt.Errorf("Could not parse 'fork': %v", err)
	}

	// iteration
	i, err := strconv.Atoi(rec[7])
	if err != nil {
		return nil, fmt.Errorf("Could not parse 'iteration': %v", err)
	}

	// value_count
	vc, err := strconv.Atoi(rec[10])
	if err != nil {
		return nil, fmt.Errorf("Could not parse 'value_count': %v", err)
	}

	// value
	v, err := strconv.ParseFloat(rec[11], 64)
	if err != nil {
		return nil, fmt.Errorf("Could not parse 'value': %v", err)
	}

	is := make([]float64, vc)
	for i := 0; i < vc; i++ {
		is[i] = v
	}

	return &InvocationsFlat{
		Benchmark:   b,
		Instance:    rec[4],
		Trial:       t,
		Fork:        f,
		Iteration:   i,
		Invocations: is,
	}, nil
}
