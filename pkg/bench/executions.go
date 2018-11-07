package bench

import (
	"fmt"
	"sync"
)

const (
	DefaultInstancesSize   = 1
	DefaultTrialsSize      = 5
	DefaultForksSize       = 10
	DefaultIterationsSize  = 20
	DefaultInvocationsSize = 1000
)

var defaultArraySizes = ArraySizes{
	Instances:   DefaultInstancesSize,
	Trials:      DefaultTrialsSize,
	Forks:       DefaultForksSize,
	Iterations:  DefaultIterationsSize,
	Invocations: DefaultInvocationsSize,
}

type ArraySizes struct {
	Instances   int
	Trials      int
	Forks       int
	Iterations  int
	Invocations int
}

type InvocationsFlat struct {
	Benchmark   B
	Instance    string
	Trial       int
	Fork        int
	Iteration   int
	Invocations Invocations
}

type Invocations []float64

type Iteration Invocations

type Fork []Iteration

type Trial []Fork

type InstanceID string

func NewInstanceID(id string) InstanceID {
	return InstanceID(id)
}

type Instance struct {
	ID     InstanceID
	Trials []Trial
}

func NewInstance(id string) *Instance {
	return NewInstanceTrialSize(id, 10)
}

func NewInstanceTrialSize(id string, ts int) *Instance {
	iid := NewInstanceID(id)
	return &Instance{
		ID:     iid,
		Trials: make([]Trial, 0, ts),
	}
}

type Execution struct {
	Benchmark      B
	Instances      map[InstanceID]*Instance
	arraySizes     ArraySizes
	indexStartsAt1 bool
	lenLock        sync.RWMutex
	len            int
}

func NewExecution(b B) *Execution {
	return NewExecutionWithIndexAndDefaults(b, true, defaultArraySizes)
}

func NewExecutionWithIndex(b B, idxAt1 bool) *Execution {
	return NewExecutionWithIndexAndDefaults(b, idxAt1, defaultArraySizes)
}

func NewExecutionWithIndexAndDefaults(b B, idxAt1 bool, d ArraySizes) *Execution {
	return &Execution{
		Benchmark:      b,
		Instances:      make(map[InstanceID]*Instance),
		arraySizes:     d,
		indexStartsAt1: idxAt1,
	}
}

func (e *Execution) Len() int {
	e.lenLock.RLock()
	defer e.lenLock.RUnlock()
	return e.len
}

func (e *Execution) addLen(i int) {
	e.lenLock.Lock()
	defer e.lenLock.Unlock()
	e.len += i
}

func (e *Execution) AddInvocations(is InvocationsFlat) error {
	if e == nil {
		return fmt.Errorf("Execution e is nil")
	}

	if !e.Benchmark.Equals(is.Benchmark) {
		return fmt.Errorf("Execution belongs to %v, not to %v", e.Benchmark, is.Benchmark)
	}

	// assume that trial, fork, and iteration number starts at 1
	if e.indexStartsAt1 {
		is.Trial--
		is.Fork--
		is.Iteration--
	}

	iid := NewInstanceID(is.Instance)
	i, insExists := e.Instances[iid]

	// check if instance exists
	if !insExists {
		i = NewInstance(is.Instance)
		iid = i.ID
		e.Instances[iid] = i
	}

	// check if trial exists
	t := is.Trial
	lt := len(i.Trials)
	if t >= lt {
		i.Trials = append(i.Trials, make(Trial, 0, e.arraySizes.Trials))
		t = lt
	}
	trial := i.Trials[t]

	// check if fork exists
	f := is.Fork
	lf := len(trial)
	if f >= lf {
		i.Trials[t] = append(trial, make(Fork, 0, e.arraySizes.Forks))
		f = lf
		trial = i.Trials[t]
	}
	fork := trial[f]

	// check if iterations exists
	it := is.Iteration
	lit := len(fork)
	if it >= lit {
		i.Trials[t][f] = append(fork, make(Iteration, 0, e.arraySizes.Iterations))
		it = lit
		fork = i.Trials[t][f]
	}
	iteration := fork[it]

	// append invocation to iteration
	i.Trials[t][f][it] = append(iteration, is.Invocations...)
	e.addLen(len(is.Invocations))
	return nil
}
