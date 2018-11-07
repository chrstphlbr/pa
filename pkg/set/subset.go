package set

func SubSetInt(super, sub Int) bool {
	for k := range sub {
		if _, contained := super[k]; !contained {
			return false
		}
	}
	return true
}

func SubSetString(super, sub String) bool {
	for k := range sub {
		if _, contained := super[k]; !contained {
			return false
		}
	}
	return true
}
