package bench_test

import (
	"testing"

	"bitbucket.org/sealuzh/pa/pkg/bench"
)

func TestBCompareEqualName(t *testing.T) {
	b1 := bench.New("b1")

	if b1.Compare(b1) != 0 {
		t.Fatalf("Should be 0")
	}
}

func TestBCompareEqualFunctionParams(t *testing.T) {
	b1 := bench.New("b1")
	b1.FunctionParams = []string{"a", "b", "c"}

	if b1.Compare(b1) != 0 {
		t.Fatalf("Should be 0")
	}
}

func TestBCompareEqualPerfParams(t *testing.T) {
	b1 := bench.New("b1")
	b1.AddPerfParam("p1", "v1")
	b1.AddPerfParam("p2", "v2")
	b1.AddPerfParam("p3", "v3")

	if b1.Compare(b1) != 0 {
		t.Fatalf("Should be 0")
	}
}

func TestBCompareEqualAll(t *testing.T) {
	b1 := bench.New("b1")
	b1.FunctionParams = []string{"a", "b", "c"}
	b1.AddPerfParam("p1", "v1")
	b1.AddPerfParam("p2", "v2")
	b1.AddPerfParam("p3", "v3")

	if b1.Compare(b1) != 0 {
		t.Fatalf("Should be 0")
	}
}

// smaller

func TestBCompareSmallerName(t *testing.T) {
	b1 := bench.New("b1")
	b2 := bench.New("b2")
	if c := b1.Compare(b2); c != -1 {
		t.Fatalf("Expected %d, was %d", -1, c)
	}
}

func TestBCompareSmallerFunctionParamsLen(t *testing.T) {
	b1 := bench.New("b1")
	b2 := bench.New("b1")
	b2.FunctionParams = []string{"a", "b", "c"}
	if c := b1.Compare(b2); c != -1 {
		t.Fatalf("Expected %d, was %d", -1, c)
	}
}

func TestBCompareSmallerFunctionParamsElem(t *testing.T) {
	b1 := bench.New("b1")
	b1.FunctionParams = []string{"a", "a", "c"}
	b2 := bench.New("b1")
	b2.FunctionParams = []string{"a", "b", "c"}
	if c := b1.Compare(b2); c != -1 {
		t.Fatalf("Expected %d, was %d", -1, c)
	}
}

func TestBCompareSmallerPerfParamsLen(t *testing.T) {
	b1 := bench.New("b1")
	b2 := bench.New("b1")
	b2.AddPerfParam("p1", "v1")
	b2.AddPerfParam("p2", "v2")
	b2.AddPerfParam("p3", "v3")
	if c := b1.Compare(b2); c != -1 {
		t.Fatalf("Expected %d, was %d", -1, c)
	}
}

func TestBCompareSmallerPerfParamsKey(t *testing.T) {
	b1 := bench.New("b1")
	b1.AddPerfParam("p2", "v1")
	b1.AddPerfParam("p3", "v2")
	b1.AddPerfParam("p6", "v3")
	b2 := bench.New("b1")
	b2.AddPerfParam("p2", "v1")
	b2.AddPerfParam("p4", "v2")
	b2.AddPerfParam("p6", "v3")
	if c := b1.Compare(b2); c != -1 {
		t.Fatalf("Expected %d, was %d", -1, c)
	}
}

func TestBCompareSmallerPerfParamsValue(t *testing.T) {
	b1 := bench.New("b1")
	b1.AddPerfParam("p1", "v1")
	b1.AddPerfParam("p2", "v1")
	b1.AddPerfParam("p3", "v3")
	b2 := bench.New("b1")
	b2.AddPerfParam("p1", "v1")
	b2.AddPerfParam("p2", "v2")
	b2.AddPerfParam("p3", "v3")
	if c := b1.Compare(b2); c != -1 {
		t.Fatalf("Expected %d, was %d", -1, c)
	}
}

// bigger

func TestBCompareBiggerName(t *testing.T) {
	b1 := bench.New("b1")
	b2 := bench.New("b2")
	if c := b2.Compare(b1); c != 1 {
		t.Fatalf("Expected %d, was %d", 1, c)
	}
}

func TestBCompareBiggerFunctionParamsLen(t *testing.T) {
	b1 := bench.New("b1")
	b1.FunctionParams = []string{"a", "b", "c"}
	b2 := bench.New("b1")
	if c := b1.Compare(b2); c != 1 {
		t.Fatalf("Expected %d, was %d", 1, c)
	}
}

func TestBCompareBiggerFunctionParamsElem(t *testing.T) {
	b1 := bench.New("b1")
	b1.FunctionParams = []string{"a", "b", "c"}
	b2 := bench.New("b1")
	b2.FunctionParams = []string{"a", "a", "c"}
	if c := b1.Compare(b2); c != 1 {
		t.Fatalf("Expected %d, was %d", 1, c)
	}
}

func TestBCompareBiggerPerfParamsLen(t *testing.T) {
	b1 := bench.New("b1")
	b1.AddPerfParam("p1", "v1")
	b1.AddPerfParam("p2", "v2")
	b1.AddPerfParam("p3", "v3")
	b2 := bench.New("b1")
	if c := b1.Compare(b2); c != 1 {
		t.Fatalf("Expected %d, was %d", 1, c)
	}
}

func TestBCompareBiggerPerfParamsKey(t *testing.T) {
	b1 := bench.New("b1")
	b1.AddPerfParam("p2", "v1")
	b1.AddPerfParam("p4", "v2")
	b1.AddPerfParam("p6", "v3")
	b2 := bench.New("b1")
	b2.AddPerfParam("p2", "v1")
	b2.AddPerfParam("p3", "v2")
	b2.AddPerfParam("p6", "v3")
	if c := b1.Compare(b2); c != 1 {
		t.Fatalf("Expected %d, was %d", 1, c)
	}
}

func TestBCompareBiggerPerfParamsValue(t *testing.T) {
	b1 := bench.New("b1")
	b1.AddPerfParam("p1", "v1")
	b1.AddPerfParam("p2", "v2")
	b1.AddPerfParam("p3", "v3")
	b2 := bench.New("b1")
	b2.AddPerfParam("p1", "v1")
	b2.AddPerfParam("p2", "v1")
	b2.AddPerfParam("p3", "v3")
	if c := b1.Compare(b2); c != 1 {
		t.Fatalf("Expected %d, was %d", 1, c)
	}
}
