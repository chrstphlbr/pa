package stat_test

import (
	"testing"

	"bitbucket.org/sealuzh/pa/pkg/stat"
)

func TestMedianEmpty(t *testing.T) {
	var d []float64 = []float64{}
	m := stat.Median(d)
	if m != 0 {
		t.Fatal()
	}
}

func TestMedianOne(t *testing.T) {
	var e float64 = 3.0
	var d []float64 = []float64{e}
	m := stat.Median(d)
	if m != e {
		t.Fatal()
	}
}

func TestMedianEven(t *testing.T) {
	var d []float64 = []float64{1, 9, 3, 5}
	m := stat.Median(d)
	if m != 4 {
		t.Fatal()
	}
}

func TestMedianUneven(t *testing.T) {
	var d []float64 = []float64{1, 9, 3, 5, 20}
	m := stat.Median(d)
	if m != 5 {
		t.Fatal()
	}
}
