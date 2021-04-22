package bench

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type B struct {
	Name           string
	FunctionParams FunctionParams
	PerfParams     *PerfParams
}

func New(name string) *B {
	return &B{
		Name:           name,
		FunctionParams: []string{},
		PerfParams:     newPerfParams(),
	}
}

func (b *B) String() string {
	var sb strings.Builder
	sb.WriteString(b.Name)
	sb.WriteString("(")
	for i, fp := range b.FunctionParams {
		if i != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fp)
	}
	sb.WriteString(")")
	sb.WriteString("{")
	sb.WriteString(b.PerfParams.String())
	sb.WriteString("}")
	return sb.String()
}

func (b *B) Copy() *B {
	nb := New(b.Name)

	nfps := make([]string, len(b.FunctionParams))
	copy(nfps, b.FunctionParams)
	nb.FunctionParams = nfps

	nb.PerfParams = b.PerfParams.Copy()

	return nb
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

	b.PerfParams.l.RLock()
	defer b.PerfParams.l.RUnlock()
	other.PerfParams.l.RLock()
	defer other.PerfParams.l.RUnlock()

	ppKeys := b.PerfParams.keys
	oppKeys := other.PerfParams.keys
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
		bval, ok := b.PerfParams.params[bkey]
		if !ok {
			panic(fmt.Sprintf("Illegal state for %s %v: performance-parameter data structures out of sync", b.Name, b.FunctionParams))
		}
		oval, ok := other.PerfParams.params[okey]
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

type FunctionParams []string

func (fp FunctionParams) String() string {
	var sb strings.Builder

	first := true
	for _, p := range fp {
		if first {
			first = false
		} else {
			sb.WriteString(",")
		}
		sb.WriteString(p)
	}

	return sb.String()
}

type PerfParams struct {
	l      sync.RWMutex
	keys   []string
	params map[string]string
}

func newPerfParams() *PerfParams {
	return &PerfParams{
		keys:   []string{},
		params: map[string]string{},
	}
}

func (pp *PerfParams) Add(param, value string) {
	pp.l.Lock()
	defer pp.l.Unlock()
	pp.keys = append(pp.keys, param)
	sort.Strings(pp.keys)
	pp.params[param] = value
	lk := len(pp.keys)
	lp := len(pp.params)
	if lk != lp {
		panic(fmt.Sprintf("Illegal state for: performance-parameter data structures out of sync (keys=%d, values=%d)", lk, lp))
	}
}

func (pp *PerfParams) Get() map[string]string {
	pp.l.RLock()
	defer pp.l.RUnlock()
	return pp.params
}

func (pp *PerfParams) Keys() []string {
	pp.l.RLock()
	defer pp.l.RUnlock()
	return pp.keys
}

func (pp *PerfParams) String() string {
	pp.l.RLock()
	defer pp.l.RUnlock()
	var sb strings.Builder

	for i, k := range pp.keys {
		v, ok := pp.params[k]
		if !ok {
			panic(fmt.Sprintf("Illegal state for: performance-parameter data structures out of sync"))
		}

		if i != 0 {
			sb.WriteString(",")
		}

		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v)
	}

	return sb.String()
}

func (pp *PerfParams) Copy() *PerfParams {
	pp.l.RLock()
	defer pp.l.RUnlock()

	newKeys := make([]string, len(pp.keys))
	copy(newKeys, pp.keys)

	newParams := make(map[string]string)
	for k, v := range pp.params {
		newParams[k] = v
	}

	return &PerfParams{
		keys:   newKeys,
		params: newParams,
	}
}
