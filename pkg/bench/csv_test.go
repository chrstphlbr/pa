package bench_test

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/chrstphlbr/pa/pkg/bench"
)

const csvErr = "record on line 1: wrong number of fields"

func noErrValueExpected(t *testing.T, bv bench.ExecutionValue) {
	if bv.Err != nil {
		t.Fatalf("No error value expected, but provided: %v", bv.Err)
	}
}

func noExecValueExpected(t *testing.T, bv bench.ExecutionValue) {
	if bv.Exec != nil {
		t.Fatalf("No execution value expected for %v, but provided: %v", bv.Type, bv.Exec)
	}
}

func fromCSVHelper(t *testing.T, file io.Reader, expectedBenchmarks int, expectedError bool) ([]*bench.Execution, error) {
	c, err := bench.FromCSV(context.TODO(), file)
	if err != nil {
		if strings.HasSuffix(err.Error(), csvErr) {
			return nil, err
		}
		t.Fatalf("Could not get Benchmark channel: %v", err)
	}

	var started bool
	var ended bool
	var cnt int

	execs := make([]*bench.Execution, 0, expectedBenchmarks)

	for bv := range c {
		switch bv.Type {
		case bench.ExecStart:
			noErrValueExpected(t, bv)
			noExecValueExpected(t, bv)
			started = true
		case bench.ExecEnd:
			noErrValueExpected(t, bv)
			noExecValueExpected(t, bv)
			ended = true
		case bench.ExecError:
			noExecValueExpected(t, bv)
			if !expectedError {
				t.Fatalf("Unexpected error: %v", bv.Err)
			} else if bv.Err == nil {
				t.Fatalf("Expected error but no error value provided (bench.ExecutionValue.Err == nil)")
			} else {
				return nil, bv.Err
			}
		case bench.ExecNext:
			noErrValueExpected(t, bv)
			if bv.Exec == nil {
				t.Fatalf("Expected execution value but got nil")
			}
			execs = append(execs, bv.Exec)
		}
		cnt++
	}

	// expected benchmarks + start + ended
	expectedExecutionValues := expectedBenchmarks + 2
	if !started || !ended || cnt != expectedExecutionValues {
		t.Fatalf("started = %t, stopped = %t, messages = %d", started, ended, cnt)
	}

	return execs, nil
}

func header(t *testing.T) (*csv.Writer, fmt.Stringer) {
	sb := &strings.Builder{}
	w := csv.NewWriter(sb)
	w.Comma = ';'
	err := w.Write([]string{"project", "commit", "benchmark", "params", "instance", "trial", "fork", "iteration", "mode", "unit", "value_count", "value"})
	if err != nil {
		t.Fatalf("Could not write header: %v", err)
	}
	w.Flush()
	return w, sb
}

func TestFromCSVEmpty(t *testing.T) {
	sr := strings.NewReader("")
	es, err := fromCSVHelper(t, sr, 0, false)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(es) != 0 {
		t.Fatalf("Expected 0 execution values")
	}
}

func TestFromCSVInvalidColumnCount(t *testing.T) {
	sr := strings.NewReader("a;b;c;")
	_, err := fromCSVHelper(t, sr, 0, false)
	if err == nil {
		t.Fatalf("Expected error %v", err)
	}

	if !strings.HasSuffix(err.Error(), csvErr) {
		t.Fatalf("Invalid error; expected to end in '%s', but was '%s'", csvErr, err.Error())
	}
}

func TestFromCSVHeaderAndComma(t *testing.T) {
	_, sb := header(t)
	sr := strings.NewReader(sb.String())
	es, err := fromCSVHelper(t, sr, 0, false)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(es) != 0 {
		t.Fatalf("Expected 0 execution values")
	}
}

func TestFromCSVPerfParams(t *testing.T) {
	w, sb := header(t)
	w.Write([]string{"", "", "b1", "p1=1", "i1", "1", "1", "1", "sample", "ms/op", "1", "1.0"})
	w.Write([]string{"", "", "b1", "p1=1,p2=2", "i1", "1", "1", "1", "sample", "ms/op", "1", "1.0"})
	w.Write([]string{"", "", "b1", "p1=1,1,p2=2,1", "i1", "1", "1", "1", "sample", "ms/op", "1", "1.0"})
	w.Write([]string{"", "", "b1", "p1=1,1,p2=2,1,p3=3,1", "i1", "1", "1", "1", "sample", "ms/op", "1", "1.0"})
	w.Flush()
	sr := strings.NewReader(sb.String())
	es, err := fromCSVHelper(t, sr, 4, false)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if l := len(es); l != 4 {
		t.Fatalf("Expected 4 execution values, got %d", l)
	}

	ex1 := es[0]
	ex1PP := ex1.Benchmark.PerfParams
	if l := len(ex1PP.Keys()); l != 1 {
		t.Fatalf("Expected 1 perf param, got %d", l)
	}
	ex1p1val, ok := ex1PP.Get()["p1"]
	if !ok {
		t.Fatalf("Expected value for perf param '%s'", "p1")
	}
	if ex1p1val != "1" {
		t.Fatalf("Expected perf-param value %s for '%s'", "1", "p1")
	}

	ex2 := es[1]
	ex2PP := ex2.Benchmark.PerfParams
	if l := len(ex2PP.Keys()); l != 2 {
		t.Fatalf("Expected 2 perf param, got %d", l)
	}

	ex2p1val, ok := ex2PP.Get()["p1"]
	if !ok {
		t.Fatalf("Expected value for perf param '%s'", "p1")
	}
	if ex2p1val != "1" {
		t.Fatalf("Expected perf-param value %s for '%s'", "1", "p1")
	}
	ex2p2val, ok := ex2PP.Get()["p2"]
	if !ok {
		t.Fatalf("Expected value for perf param '%s'", "p2")
	}
	if ex2p2val != "2" {
		t.Fatalf("Expected perf-param value %s for '%s'", "2", "p2")
	}

	ex3 := es[2]
	ex3PP := ex3.Benchmark.PerfParams
	if l := len(ex3PP.Keys()); l != 2 {
		t.Fatalf("Expected 2 perf param, got %d", l)
	}

	ex3p1val, ok := ex3PP.Get()["p1"]
	if !ok {
		t.Fatalf("Expected value for perf param '%s'", "p1")
	}
	if ex3p1val != "1,1" {
		t.Fatalf("Expected perf-param value %s for '%s'", "1,1", "p1")
	}
	ex3p2val, ok := ex3PP.Get()["p2"]
	if !ok {
		t.Fatalf("Expected value for perf param '%s'", "p2")
	}
	if ex3p2val != "2,1" {
		t.Fatalf("Expected perf-param value %s for '%s'", "2,1", "p2")
	}

	ex4 := es[3]
	ex4PP := ex4.Benchmark.PerfParams
	if l := len(ex4PP.Keys()); l != 3 {
		t.Fatalf("Expected 3 perf param, got %d", l)
	}

	ex4p1val, ok := ex4PP.Get()["p1"]
	if !ok {
		t.Fatalf("Expected value for perf param '%s'", "p1")
	}
	if ex4p1val != "1,1" {
		t.Fatalf("Expected perf-param value %s for '%s'", "1,1", "p1")
	}
	ex4p2val, ok := ex4PP.Get()["p2"]
	if !ok {
		t.Fatalf("Expected value for perf param '%s'", "p2")
	}
	if ex4p2val != "2,1" {
		t.Fatalf("Expected perf-param value %s for '%s'", "2,1", "p2")
	}
	ex4p3val, ok := ex4PP.Get()["p3"]
	if !ok {
		t.Fatalf("Expected value for perf param '%s'", "p3")
	}
	if ex4p3val != "3,1" {
		t.Fatalf("Expected perf-param value %s for '%s'", "3,1", "p3")
	}
}

func TestFromCSVSingleInvs(t *testing.T) {
	w, sb := header(t)

	//[]string{"project", "commit", "benchmark", "params", "instance", "trial", "fork", "iteration", "mode", "unit", "value_count", "value"}
	err := w.Write([]string{"", "", "b1", "", "i1", "1", "1", "1", "", "", "1", "0.0"})
	if err != nil {
		t.Fatalf("Could not write to CSV: %v", err)
	}
	w.Flush()

	sr := strings.NewReader(sb.String())
	es, err := fromCSVHelper(t, sr, 1, false)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(es) != 1 {
		t.Fatalf("Expected 1 element, but was %d", len(es))
	}

	checkInstance(t, es[0], "i1", 1, 1, 1, 1, 1, 1)
}

func writeInvocations(t *testing.T, w *csv.Writer, b *bench.B, ins string, trial, fork, iter, nrinvs int) {
	invs := createInvocationsFlat(nrinvs, b, ins, trial, fork, iter)
	for _, iv := range invs {
		//[]string{"project", "commit", "benchmark", "params", "instance", "trial", "fork", "iteration", "mode", "unit", "value_count", "value"}
		err := w.Write([]string{"", "", iv.Benchmark.Name, "", iv.Instance, strconv.Itoa(iv.Trial), strconv.Itoa(iv.Fork), strconv.Itoa(iv.Iteration), "", "", strconv.Itoa(iv.Invocations.Count), strconv.FormatFloat(iv.Invocations.Value, 'e', -1, 64)})
		if err != nil {
			t.Fatalf("Could not write to CSV: %v", err)
		}
		w.Flush()
	}
}

type replicationPoint int

const (
	rpIns replicationPoint = iota
	rpTrial
	rpFork
	rpIter
	rpInv
)

func fromCSVMultiInvs(t *testing.T, nrbs, nrins, nrtrials, nrforks, nriters, nrinvs, invcount int) {
	w, sb := header(t)

	benchsChecked := map[string]bool{} // key is bench.B.Name
	ins := []string{}

	// write CSV
	for b := 1; b <= nrbs; b++ {
		bench := bench.New(fmt.Sprintf("b%d", b))
		benchsChecked[bench.Name] = false
		for i := 1; i <= nrins; i++ {
			instance := fmt.Sprintf("i%d", i)
			ins = append(ins, instance)
			for tr := 1; tr <= nrtrials; tr++ {
				for f := 1; f <= nrforks; f++ {
					for it := 1; it <= nriters; it++ {
						for inv := 1; inv <= nrinvs; inv++ {
							writeInvocations(t, w, bench, instance, tr, f, it, invcount)
						}
					}
				}
			}
		}
	}

	sr := strings.NewReader(sb.String())
	es, err := fromCSVHelper(t, sr, nrbs, false)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// printExecutions(es)

	for _, e := range es {
		_, shouldBeThere := benchsChecked[e.Benchmark.Name]
		if shouldBeThere {
			benchsChecked[e.Benchmark.Name] = true
		} else {
			t.Fatalf("Have results for benchmark (%s) that should not be there", e.Benchmark.Name)
		}

		if len(e.Instances) != nrins {
			t.Fatalf("Invalid number of instances: was %d, expected %d", len(e.Instances), nrins)
		}
		for _, insName := range ins {
			checkInstance(t, e, insName, invcount, nrins, nrtrials, nrforks, nriters, nrinvs*invcount)
		}
	}

	for b, checked := range benchsChecked {
		if !checked {
			t.Fatalf("%+v not in executions", b)
		}
	}
}

func TestFromCSVMultiInvs1(t *testing.T) {
	fromCSVMultiInvs(t, 1, 1, 1, 1, 1, 1, 20)
}

func TestFromCSVMultiInvs2(t *testing.T) {
	fromCSVMultiInvs(t, 1, 1, 1, 1, 1, 5, 20)
}

func TestFromCSVMultiInvsIter(t *testing.T) {
	fromCSVMultiInvs(t, 1, 1, 1, 1, 5, 1, 20)
}

func TestFromCSVMultiInvsForks(t *testing.T) {
	fromCSVMultiInvs(t, 1, 1, 1, 5, 1, 1, 20)
}

func TestFromCSVMultiInvsTrials(t *testing.T) {
	fromCSVMultiInvs(t, 1, 1, 5, 1, 1, 1, 20)
}

func TestFromCSVMultiInvsIns(t *testing.T) {
	fromCSVMultiInvs(t, 1, 5, 1, 1, 1, 1, 20)
}

func TestFromCSVMultiInvsBenchs(t *testing.T) {
	fromCSVMultiInvs(t, 5, 1, 1, 1, 1, 1, 20)
}

func TestFromCSVMultiInvsAll(t *testing.T) {
	fromCSVMultiInvs(t, 5, 5, 5, 5, 5, 5, 20)
}
