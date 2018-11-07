package bench

type B struct {
	Name           string
	FunctionParams []string
	PerfParams     map[string]string
}

func (b B) Equals(other B) bool {
	if b.Name != other.Name {
		return false
	}

	// function parameters
	if len(b.FunctionParams) != len(other.FunctionParams) {
		return false
	}

	for i := 0; i < len(b.FunctionParams); i++ {
		if b.FunctionParams[i] != other.FunctionParams[i] {
			return false
		}
	}

	// performance-test parameters
	if len(b.PerfParams) != len(other.PerfParams) {
		return false
	}

	for k, v := range b.PerfParams {
		otherv, exists := other.PerfParams[k]
		if !exists {
			// key does not exist in other.PerfParam
			return false
		} else if v != otherv {
			// values do not match
			return false
		}
	}

	return true
}
