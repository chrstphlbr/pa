package bench_test

import (
	"testing"

	"bitbucket.org/sealuzh/pa/pkg/bench"
)

func TestBEqualsUnset(t *testing.T) {
	b1 := bench.B{}
	b2 := bench.B{}

	if !b1.Equals(b2) {
		t.Fail()
	}
}

// Name

func TestBEqualsB1NameB2Unset(t *testing.T) {
	b1 := bench.B{
		Name: "b1",
	}
	b2 := bench.B{}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1UnsetB2Name(t *testing.T) {
	b1 := bench.B{}
	b2 := bench.B{
		Name: "b2",
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1NameB2NameEqual(t *testing.T) {
	b1 := bench.B{
		Name: "b1",
	}
	b2 := bench.B{
		Name: "b1",
	}

	if !b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1NameB2NameUnequal(t *testing.T) {
	b1 := bench.B{
		Name: "b1",
	}
	b2 := bench.B{
		Name: "b2",
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

// FunctionParams

func TestBEqualsB1FPB2Unset(t *testing.T) {
	b1 := bench.B{
		FunctionParams: []string{"p1"},
	}
	b2 := bench.B{}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1UnsetB2FP(t *testing.T) {
	b1 := bench.B{}
	b2 := bench.B{
		FunctionParams: []string{"p1"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1FPB2FPEqual(t *testing.T) {
	b1 := bench.B{
		FunctionParams: []string{"p1"},
	}
	b2 := bench.B{
		FunctionParams: []string{"p1"},
	}

	if !b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1FPB2FPUnequal(t *testing.T) {
	b1 := bench.B{
		FunctionParams: []string{"p1"},
	}
	b2 := bench.B{
		FunctionParams: []string{"p2"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1FP2B2FP2Equal(t *testing.T) {
	b1 := bench.B{
		FunctionParams: []string{"p1", "p2"},
	}
	b2 := bench.B{
		FunctionParams: []string{"p1", "p2"},
	}

	if !b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1FP2B2FP2Unequal(t *testing.T) {
	b1 := bench.B{
		FunctionParams: []string{"p1", "p2"},
	}
	b2 := bench.B{
		FunctionParams: []string{"p2", "p1"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1FP2B2FPUnequal(t *testing.T) {
	b1 := bench.B{
		FunctionParams: []string{"p1", "p2"},
	}
	b2 := bench.B{
		FunctionParams: []string{"p2"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1FPB2FP2Unequal(t *testing.T) {
	b1 := bench.B{
		FunctionParams: []string{"p1"},
	}
	b2 := bench.B{
		FunctionParams: []string{"p1", "p2"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

// PerfParams

func TestBEqualsB1PPB2Unset(t *testing.T) {
	b1 := bench.B{
		PerfParams: map[string]string{"pp1": "ppv1"},
	}
	b2 := bench.B{}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1UnsetB2PP(t *testing.T) {
	b1 := bench.B{}
	b2 := bench.B{
		PerfParams: map[string]string{"pp1": "ppv1"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1PPB2PPEqual(t *testing.T) {
	b1 := bench.B{
		PerfParams: map[string]string{"pp1": "ppv1"},
	}
	b2 := bench.B{
		PerfParams: map[string]string{"pp1": "ppv1"},
	}

	if !b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1PPB2PPUnequal(t *testing.T) {
	b1 := bench.B{
		PerfParams: map[string]string{"pp1": "ppv1"},
	}
	b2 := bench.B{
		PerfParams: map[string]string{"pp2": "ppv1"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1PP2B2PPUnequal(t *testing.T) {
	b1 := bench.B{
		PerfParams: map[string]string{"pp1": "ppv1", "pp2": "ppv1"},
	}
	b2 := bench.B{
		PerfParams: map[string]string{"pp1": "ppv1"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1PPB2PP2Unequal(t *testing.T) {
	b1 := bench.B{
		PerfParams: map[string]string{"pp1": "ppv1"},
	}
	b2 := bench.B{
		PerfParams: map[string]string{"pp1": "ppv1", "pp2": "ppv1"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1PP2B2PP2Equal(t *testing.T) {
	b1 := bench.B{
		PerfParams: map[string]string{"pp2": "ppv1", "pp1": "ppv1"},
	}
	b2 := bench.B{
		PerfParams: map[string]string{"pp1": "ppv1", "pp2": "ppv1"},
	}

	if !b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsB1PP2B2PP2Equal2(t *testing.T) {
	b1 := bench.B{
		PerfParams: map[string]string{"pp1": "ppv1", "pp2": "ppv1"},
	}
	b2 := bench.B{
		PerfParams: map[string]string{"pp2": "ppv1", "pp1": "ppv1"},
	}

	if !b1.Equals(b2) {
		t.Fail()
	}
}

// complex

func TestBEqualsSameNameSameFP(t *testing.T) {
	b1 := bench.B{
		Name:           "b1",
		FunctionParams: []string{"p1", "p2"},
	}
	b2 := bench.B{
		Name:           "b1",
		FunctionParams: []string{"p1", "p2"},
	}

	if !b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsSameNameDifferentFP(t *testing.T) {
	b1 := bench.B{
		Name:           "b1",
		FunctionParams: []string{"p1", "p2"},
	}
	b2 := bench.B{
		Name:           "b1",
		FunctionParams: []string{"p1"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsSameNameSamePP(t *testing.T) {
	b1 := bench.B{
		Name:       "b1",
		PerfParams: map[string]string{"pp1": "ppv1", "pp2": "ppv1"},
	}
	b2 := bench.B{
		Name:       "b1",
		PerfParams: map[string]string{"pp1": "ppv1", "pp2": "ppv1"},
	}

	if !b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsSameNameDifferentPP(t *testing.T) {
	b1 := bench.B{
		Name:       "b1",
		PerfParams: map[string]string{"pp1": "ppv1", "pp2": "ppv1"},
	}
	b2 := bench.B{
		Name:       "b1",
		PerfParams: map[string]string{"pp1": "ppv1"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsDifferntNameSameFP(t *testing.T) {
	b1 := bench.B{
		Name:           "b1",
		FunctionParams: []string{"p1", "p2"},
	}
	b2 := bench.B{
		Name:           "b2",
		FunctionParams: []string{"p1", "p2"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsDifferentNameSamePP(t *testing.T) {
	b1 := bench.B{
		Name:       "b1",
		PerfParams: map[string]string{"pp1": "ppv1"},
	}
	b2 := bench.B{
		Name:       "b2",
		PerfParams: map[string]string{"pp1": "ppv1"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsSameNameSameFPSamePP(t *testing.T) {
	b1 := bench.B{
		Name:           "b1",
		FunctionParams: []string{"p1", "p2"},
		PerfParams:     map[string]string{"pp1": "ppv1"},
	}
	b2 := bench.B{
		Name:           "b1",
		FunctionParams: []string{"p1", "p2"},
		PerfParams:     map[string]string{"pp1": "ppv1"},
	}

	if !b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsSameNameSameFPDifferentPP(t *testing.T) {
	b1 := bench.B{
		Name:           "b1",
		FunctionParams: []string{"p1", "p2"},
		PerfParams:     map[string]string{"pp1": "ppv1"},
	}
	b2 := bench.B{
		Name:           "b1",
		FunctionParams: []string{"p1", "p2"},
		PerfParams:     map[string]string{"pp2": "ppv1"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsSameNameDifferentFPSamePP(t *testing.T) {
	b1 := bench.B{
		Name:           "b1",
		FunctionParams: []string{"p1"},
		PerfParams:     map[string]string{"pp1": "ppv1"},
	}
	b2 := bench.B{
		Name:           "b1",
		FunctionParams: []string{"p1", "p2"},
		PerfParams:     map[string]string{"pp1": "ppv1"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}

func TestBEqualsDifferentNameSameFPSamePP(t *testing.T) {
	b1 := bench.B{
		Name:           "b1",
		FunctionParams: []string{"p1", "p2"},
		PerfParams:     map[string]string{"pp1": "ppv1"},
	}
	b2 := bench.B{
		Name:           "b2",
		FunctionParams: []string{"p1", "p2"},
		PerfParams:     map[string]string{"pp1": "ppv1"},
	}

	if b1.Equals(b2) {
		t.Fail()
	}
}
