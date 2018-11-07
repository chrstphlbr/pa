package set

func EqualInts(tg, cg Int) bool {
	if len(tg) != len(cg) {
		return false
	}

	for k := range tg {
		if _, exists := cg[k]; !exists {
			return false
		}
	}
	return true
}

func EqualStrings(tg, cg String) bool {
	if len(tg) != len(cg) {
		return false
	}

	for k := range tg {
		if _, exists := cg[k]; !exists {
			return false
		}
	}
	return true
}
