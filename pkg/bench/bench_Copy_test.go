package bench_test

import (
	"testing"

	"github.com/chrstphlbr/pa/pkg/bench"
)

func TestBCopyDifferentPointer(t *testing.T) {
	b := bench.New("bench1")
	bc := b.Copy()

	if b == bc {
		t.Fatalf("copy and original are identical")
	}

	if !b.Equals(bc) {
		t.Fatalf("copy and original not equal")
	}
}

func TestBCopyEmpty(t *testing.T) {
	b := bench.New("bench1")
	bc := b.Copy()

	if b.Name != bc.Name {
		t.Fatalf("copy's name differnt to original's name")
	}

	lfpso := len(b.FunctionParams)
	lfpsc := len(bc.FunctionParams)
	if lfpsc != 0 {
		t.Fatalf("copy's FunctionParams length not 0, was %d", lfpsc)
	}
	if lfpso != lfpsc {
		t.Fatalf("copy's FunctionParams length not equal to original's")
	}

	ppso := b.PerfParams.Get()
	lppso := len(ppso)
	ppkso := b.PerfParams.Keys()
	lppkso := len(ppkso)
	ppsc := bc.PerfParams.Get()
	lppsc := len(ppsc)
	ppksc := bc.PerfParams.Keys()
	lppksc := len(ppksc)
	if lppsc != 0 || lppksc != 0 {
		t.Fatalf("copy's PerfParams length not 0, was %d (Get()) and %d (Keys())", lppsc, lppksc)
	}
	if lppso != lppsc && lppkso != lppksc {
		t.Fatalf("copy's PerfParams length not equal to original's")
	}

	if !b.Equals(bc) {
		t.Fatalf("copy and original not equal")
	}
}

func TestBCopyComplex(t *testing.T) {
	b := bench.New("bench1")
	b.FunctionParams = []string{"a", "b", "c"}
	b.PerfParams.Add("p1", "v1")
	b.PerfParams.Add("p2", "v2")
	b.PerfParams.Add("p3", "v3")

	bc := b.Copy()

	if b.Name != bc.Name {
		t.Fatalf("copy's name differnt to original's name")
	}

	// function parameters

	expectedFunctionParams := 3

	lfpso := len(b.FunctionParams)
	lfpsc := len(bc.FunctionParams)
	if lfpsc != expectedFunctionParams {
		t.Fatalf("copy's FunctionParams length not %d, was %d", expectedFunctionParams, lfpsc)
	}
	if lfpso != lfpsc {
		t.Fatalf("copy's FunctionParams length not equal to original's")
	}

	for i, fpo := range b.FunctionParams {
		fpc := bc.FunctionParams[i]
		if fpo != fpc {
			t.Fatalf("FunctionParam at position %d not equal", i)
		}
	}

	// benchmark parameters

	expectedPerfParams := 3

	ppso := b.PerfParams.Get()
	lppso := len(ppso)
	ppkso := b.PerfParams.Keys()
	lppkso := len(ppkso)
	ppsc := bc.PerfParams.Get()
	lppsc := len(ppsc)
	ppksc := bc.PerfParams.Keys()
	lppksc := len(ppksc)
	if lppsc != expectedPerfParams || lppksc != expectedPerfParams {
		t.Fatalf("copy's PerfParams length not %d, was %d (Get()) and %d (Keys())", expectedPerfParams, lppkso, lppksc)
	}
	if lppso != lppsc && lppkso != lppksc {
		t.Fatalf("copy's PerfParams length not equal to original's")
	}

	for i, keyo := range ppkso {
		keyc := ppksc[i]
		if keyo != keyc {
			t.Fatalf("PerfParam key at position %d not equal", i)
		}

		valueo := ppso[keyo]
		valuec := ppsc[keyc]
		if valueo != valuec {
			t.Fatalf("PerfParam value for key '%s' not equal", keyo)
		}
	}

	// equals

	if !b.Equals(bc) {
		t.Fatalf("copy and original not equal")
	}
}
