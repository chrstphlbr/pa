package bench

import (
	"fmt"
	"sort"
	"sync"
)

type B struct {
	Name           string
	FunctionParams []string
	perfParamLock  sync.RWMutex
	perfParamKeys  []string
	perfParams     map[string]string
}

func New(name string) *B {
	return &B{
		Name:           name,
		FunctionParams: []string{},
		perfParamKeys:  []string{},
		perfParams:     map[string]string{},
	}
}

func (b *B) AddPerfParam(param, value string) {
	b.perfParamLock.Lock()
	defer b.perfParamLock.Unlock()
	b.perfParamKeys = append(b.perfParamKeys, param)
	sort.Strings(b.perfParamKeys)
	b.perfParams[param] = value
	if len(b.perfParamKeys) != len(b.perfParams) {
		panic(fmt.Sprintf("Illegal state for %s %v: performance-parameter data structures out of sync", b.Name, b.FunctionParams))
	}
}

func (b *B) PerfParam(param string) (string, bool) {
	b.perfParamLock.RLock()
	defer b.perfParamLock.RUnlock()
	v, ok := b.perfParams[param]
	return v, ok
}

func (b *B) PerfParams() map[string]string {
	b.perfParamLock.RLock()
	defer b.perfParamLock.RUnlock()
	return b.perfParams
}

func (b *B) PerfParamKeys() []string {
	b.perfParamLock.RLock()
	defer b.perfParamLock.RUnlock()
	return b.perfParamKeys
}

func (b *B) Equals(other *B) bool {
	return b.Compare(other) == 0
}

// Compare compares two Benchmarks lexically and returns -1 (this is smaller), 0 (they are equal), or 1 (other is smaller)
func (b *B) Compare(other *B) int {
	if b.Name < other.Name {
		return -1
	} else if b.Name > other.Name {
		return 1
	}

	// if names are equal

	// check FunctionParams
	lfp := len(b.FunctionParams)
	lofp := len(other.FunctionParams)
	if lfp < lofp {
		return -1
	} else if lfp > lofp {
		return 1
	}

	// lengths of FunctionParams are equal

	for i := range b.FunctionParams {
		bfp := b.FunctionParams[i]
		ofp := other.FunctionParams[i]
		if bfp < ofp {
			return -1
		} else if bfp > ofp {
			return 1
		}
	}

	// FunctionParams are equal

	b.perfParamLock.RLock()
	defer b.perfParamLock.RUnlock()
	other.perfParamLock.RLock()
	defer other.perfParamLock.RUnlock()

	ppKeys := b.perfParamKeys
	oppKeys := other.perfParamKeys
	lpp := len(ppKeys)
	lopp := len(oppKeys)
	if lpp < lopp {
		return -1
	} else if lpp > lopp {
		return 1
	}

	// lengths of PerfParams are equal
	for i, bkey := range ppKeys {
		// check key order
		okey := oppKeys[i]
		if bkey < okey {
			return -1
		} else if bkey > okey {
			return 1
		}

		// compare values
		bval, ok := b.perfParams[bkey]
		if !ok {
			panic(fmt.Sprintf("Illegal state for %s %v: performance-parameter data structures out of sync", b.Name, b.FunctionParams))
		}
		oval, ok := other.perfParams[okey]
		if !ok {
			panic(fmt.Sprintf("Illegal state for %s %v: performance-parameter data structures out of sync", other.Name, other.FunctionParams))
		}

		if bval < oval {
			return -1
		} else if bval > oval {
			return 1
		}
	}

	return 0
}
