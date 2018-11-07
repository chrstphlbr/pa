package set

func DisjointInts(tg, cg Int) bool {
	for k := range tg {
		if _, exists := cg[k]; exists {
			return false
		}
	}
	return true
}

func DisjointStrings(tg, cg String) bool {
	for k := range tg {
		if _, exists := cg[k]; exists {
			return false
		}
	}
	return true
}
