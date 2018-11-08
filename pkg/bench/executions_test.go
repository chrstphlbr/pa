package bench_test

import (
	"math"
	"testing"

	"bitbucket.org/sealuzh/pa/pkg/bench"
)

func createBench(name string) *bench.B {
	return bench.New(name)
}

var b = createBench("benchA")

var b2 = createBench("benchB")

func TestNewExecutionSameBenchmark(t *testing.T) {
	e := bench.NewExecution(b)
	if !b.Equals(e.Benchmark) {
		t.Fatal()
	}

	if !e.Benchmark.Equals(b) {
		t.Fatal()
	}
}

func TestNewExecutionDifferentBenchmark(t *testing.T) {
	e := bench.NewExecution(b)
	if b2.Equals(e.Benchmark) {
		t.Fatal()
	}

	if e.Benchmark.Equals(b2) {
		t.Fatal()
	}
}

func TestNewExecutionEmpty(t *testing.T) {
	e := bench.NewExecution(b)
	if len(e.Instances) != 0 {
		t.Fatal()
	}
}

// AddInvocations

func TestAddInvocationsInvalidBench(t *testing.T) {
	e := bench.NewExecution(b)

	is := createInvocationsFlat(20, b2, "", 1, 1, 1)

	err := e.AddInvocations(is)
	if err == nil {
		t.Fatal()
	}
}

func TestAddInvocationsFirst(t *testing.T) {
	e := bench.NewExecutionWithIndex(b, true)
	in := "instance1"

	nrivs := 20

	is := createInvocationsFlat(nrivs, b, in, 1, 1, 1)

	err := e.AddInvocations(is)
	if err != nil {
		t.Fatalf("Could not add: %v", err)
	}

	iid := bench.NewInstanceID(in)
	i, ie := e.Instances[iid]
	if !ie {
		t.Fatalf("Instance '%s' does not exist", in)
	}

	lt := len(i.Trials)
	if lt != 1 {
		t.Fatalf("Trial length not 1, was %d", lt)
	}

	forks := i.Trials[0]
	lf := len(forks)
	if lf != 1 {
		t.Fatalf("Forks length not 1, was %d", lf)
	}

	iterations := i.Trials[0][0]
	li := len(iterations)
	if li != 1 {
		t.Fatalf("Iterations length not 1, was %d", li)
	}

	invocations := i.Trials[0][0][0]
	livs := len(invocations)
	if livs != nrivs {
		t.Fatalf("Invocations length not %d, was %d", nrivs, livs)
	}

	for i, v := range invocations {
		expected := float64(i)
		if v != expected {
			t.Fatalf("Unexepected invocations value %f, expected %f", v, expected)
		}
	}
}

func addInvocation(t *testing.T, ins1, ins2 string, t1, t2, f1, f2, i1, i2 int) {
	e := bench.NewExecutionWithIndex(b, true)
	nrivs := 20
	is := createInvocationsFlat(nrivs, b, ins1, t1, f1, i1)

	err := e.AddInvocations(is)
	if err != nil {
		t.Fatal()
	}

	is = createInvocationsFlat(nrivs, b, ins2, t2, f2, i2)
	err = e.AddInvocations(is)
	if err != nil {
		t.Fatal()
	}

	doubleInvs := true

	enrins := 0.0
	if ins1 == ins2 {
		// same instance
		enrins = 1
	} else {
		// different instance
		enrins = 2
		doubleInvs = false
	}

	enrt := 0.0
	if t1 == t2 {
		// same trial
		enrt = 1
	} else {
		// different trial
		enrt = 2
		doubleInvs = false
	}
	enrt = math.Ceil(enrt / enrins)

	enrf := 0.0
	if f1 == f2 {
		// same fork
		enrf = 1
	} else {
		// different fork
		enrf = 2
		doubleInvs = false
	}
	enrf = math.Ceil(enrf / enrt)

	enri := 0.0
	if i1 == i2 {
		// same iteration
		enri = 1
	} else {
		enri = 2
		doubleInvs = false
	}
	enri = math.Ceil(enri / enrf)

	enrinvs := nrivs
	if doubleInvs {
		enrinvs *= 2
	}

	checkInstance(t, e, ins1, nrivs, int(enrins), int(enrt), int(enrf), int(enri), enrinvs)
	checkInstance(t, e, ins2, nrivs, int(enrins), int(enrt), int(enrf), int(enri), enrinvs)
}

func TestAddInvocationsAppendInvocationsInstance(t *testing.T) {
	addInvocation(t, "i1", "i2", 1, 1, 1, 1, 1, 1)
}

func TestAddInvocationsAppendInvocationsTrial(t *testing.T) {
	addInvocation(t, "i1", "i1", 1, 2, 1, 1, 1, 1)
}

func TestAddInvocationsAppendInvocationsFork(t *testing.T) {
	addInvocation(t, "i1", "i1", 1, 1, 1, 2, 1, 1)
}

func TestAddInvocationsAppendInvocationsIteration(t *testing.T) {
	addInvocation(t, "i1", "i1", 1, 1, 1, 1, 1, 2)
}

func TestAddInvocationsAppendInvocationsInvocation(t *testing.T) {
	addInvocation(t, "i1", "i1", 1, 1, 1, 1, 1, 1)
}
