package set

func UnionInt(a, b Int) Int {
	u := make(Int)
	for k := range a {
		u[k] = struct{}{}
	}
	for k := range b {
		u[k] = struct{}{}
	}
	return u
}

func UnionString(a, b String) String {
	u := make(String)
	for k := range a {
		u[k] = struct{}{}
	}
	for k := range b {
		u[k] = struct{}{}
	}
	return u
}
