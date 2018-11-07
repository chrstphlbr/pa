package set

import "strconv"

func NewInt(d []int) Int {
	m := make(Int)
	for _, dp := range d {
		m[dp] = struct{}{}
	}
	return m
}

// NewIntFromTo creates a set with values ranging from (inclusive) to (exclusive)
func NewIntFromTo(from, to int) Int {
	m := make(Int)
	for i := from; i < to; i++ {
		m[i] = struct{}{}
	}
	return m
}

func NewString(d []string) String {
	m := make(String)
	for _, dp := range d {
		m[dp] = struct{}{}
	}
	return m
}

func NewStringFromTo(from, to int) String {
	m := make(String)
	for i := from; i < to; i++ {
		m[strconv.Itoa(i)] = struct{}{}
	}
	return m
}
