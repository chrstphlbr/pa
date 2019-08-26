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
	DefaultInvocationsSize = 200
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
	Benchmark   *B
	Instance    string
	Trial       int
	Fork        int
	Iteration   int
	Invocations Invocations
}

type ExecutionSlice interface {
	Slice(InvocationSampler) [][][][][]float64
	FlatSlice(InvocationSampler) []float64
	// returns the length of all leaf elements
	// ElementCount() int
}

type Invocations struct {
	Count int
	Value float64
}

type Iteration struct {
	ID          int
	Invocations []Invocations
}

func (i *Iteration) Merge(other *Iteration) error {
	if i.ID != other.ID {
		return fmt.Errorf("Iteration IDs do not match: %d != %d", i.ID, other.ID)
	}
	i.Invocations = append(i.Invocations, other.Invocations...)
	return nil
}

type Fork struct {
	ID           int
	IterationIDs []int
	Iterations   map[int]*Iteration
}

func (f *Fork) Merge(other *Fork) error {
	if f.ID != other.ID {
		return fmt.Errorf("Fork IDs do not match: %d != %d", f.ID, other.ID)
	}

	for _, oiid := range other.IterationIDs {
		oi, ok := other.Iterations[oiid]
		if !ok {
			panic(fmt.Sprintf("Invalid state: IterationIDs and Iterations out of sync for %d", oiid))
		}
		i, ok := f.Iterations[oiid]
		if !ok {
			f.IterationIDs = append(f.IterationIDs, oiid)
			f.Iterations[oiid] = oi
		} else {
			err := i.Merge(oi)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type Trial struct {
	ID      int
	ForkIDs []int
	Forks   map[int]*Fork
}

func (t *Trial) Merge(other *Trial) error {
	if t.ID != other.ID {
		return fmt.Errorf("Trial IDs do not match: %d != %d", t.ID, other.ID)
	}

	for _, ofid := range other.ForkIDs {
		of, ok := other.Forks[ofid]
		if !ok {
			panic(fmt.Sprintf("Invalid state: ForkIDs and Forks out of sync for %d", ofid))
		}
		f, ok := t.Forks[ofid]
		if !ok {
			t.ForkIDs = append(t.ForkIDs, ofid)
			t.Forks[ofid] = of
		} else {
			err := f.Merge(of)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type Instance struct {
	ID       string
	TrialIDs []int
	Trials   map[int]*Trial
}

func (i *Instance) Merge(other *Instance) error {
	if i.ID != other.ID {
		return fmt.Errorf("Instance IDs do not match: %s != %s", i.ID, other.ID)
	}

	for _, otid := range other.TrialIDs {
		ot, ok := other.Trials[otid]
		if !ok {
			panic(fmt.Sprintf("Invalid state: TrialIDs and Trials out of sync for %d", otid))
		}
		t, ok := i.Trials[otid]
		if !ok {
			i.TrialIDs = append(i.TrialIDs, otid)
			i.Trials[otid] = ot
		} else {
			err := t.Merge(ot)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

var _ ExecutionSlice = &Execution{}

type Execution struct {
	Benchmark   *B
	InstanceIDs []string
	Instances   map[string]*Instance
	arraySizes  ArraySizes
	lenLock     sync.RWMutex
	len         int
}

func NewExecution(b *B) *Execution {
	return NewExecutionWithDefaults(b, defaultArraySizes)
}

func NewExecutionWithDefaults(b *B, d ArraySizes) *Execution {
	return &Execution{
		Benchmark:  b,
		Instances:  make(map[string]*Instance),
		arraySizes: d,
	}
}

func NewExecutionFromInvocationsFlat(ivf InvocationsFlat) *Execution {
	e := NewExecution(ivf.Benchmark)

	iid := ivf.Instance
	tid := ivf.Trial
	fid := ivf.Fork
	itid := ivf.Iteration

	e.InstanceIDs = append(e.InstanceIDs, iid)
	e.Instances[iid] = &Instance{
		ID:       iid,
		TrialIDs: append(make([]int, 0, e.arraySizes.Trials), tid),
		Trials: map[int]*Trial{
			tid: {
				ID:      tid,
				ForkIDs: append(make([]int, 0, e.arraySizes.Forks), fid),
				Forks: map[int]*Fork{
					fid: {
						ID:           fid,
						IterationIDs: append(make([]int, 0, e.arraySizes.Iterations), itid),
						Iterations: map[int]*Iteration{
							itid: {
								ID:          itid,
								Invocations: []Invocations{ivf.Invocations},
							},
						},
					},
				},
			},
		},
	}

	e.len = ivf.Invocations.Count

	return e
}

func (e *Execution) ElementCount() int {
	e.lenLock.RLock()
	defer e.lenLock.RUnlock()
	return e.len
}

func (e *Execution) addLen(i int) {
	e.lenLock.Lock()
	defer e.lenLock.Unlock()
	e.len += i
}

func (e *Execution) Slice(s InvocationSampler) [][][][][]float64 {
	out := make([][][][][]float64, 0, len(e.InstanceIDs))

	for _, iid := range e.InstanceIDs {
		i, ok := e.Instances[iid]
		if !ok {
			panic(fmt.Sprintf("Invalid state: InstanceIDs and Instances out of sync for %s", iid))
		}

		trialSlice := make([][][][]float64, 0, len(i.Trials))

		for _, tid := range i.TrialIDs {
			t, ok := i.Trials[tid]
			if !ok {
				panic(fmt.Sprintf("Invalid state: TrialIDs and Trials out of sync for %d", tid))
			}

			forkSlice := make([][][]float64, 0, len(t.Forks))

			for _, fid := range t.ForkIDs {
				f, ok := t.Forks[fid]
				if !ok {
					panic(fmt.Sprintf("Invalid state: ForkIDs and Forks out of sync for %d", fid))
				}

				iterationSlice := make([][]float64, 0, len(f.Iterations))

				for _, itid := range f.IterationIDs {
					it, ok := f.Iterations[itid]
					if !ok {
						panic(fmt.Sprintf("Invalid state: IterationIDs and Iterations out of sync for %d", itid))
					}
					iterationSlice = append(iterationSlice, s(it.Invocations))
				}
				forkSlice = append(forkSlice, iterationSlice)
			}
			trialSlice = append(trialSlice, forkSlice)
		}
		out = append(out, trialSlice)
	}

	return out
}

func (e *Execution) FlatSlice(s InvocationSampler) []float64 {
	out := make([]float64, 0, len(e.InstanceIDs))

	for _, iid := range e.InstanceIDs {
		i, ok := e.Instances[iid]
		if !ok {
			panic(fmt.Sprintf("Invalid state: InstanceIDs and Instances out of sync for %s", iid))
		}

		for _, tid := range i.TrialIDs {
			t, ok := i.Trials[tid]
			if !ok {
				panic(fmt.Sprintf("Invalid state: TrialIDs and Trials out of sync for %d", tid))
			}

			for _, fid := range t.ForkIDs {
				f, ok := t.Forks[fid]
				if !ok {
					panic(fmt.Sprintf("Invalid state: ForkIDs and Forks out of sync for %d", fid))
				}

				for _, itid := range f.IterationIDs {
					it, ok := f.Iterations[itid]
					if !ok {
						panic(fmt.Sprintf("Invalid state: IterationIDs and Iterations out of sync for %d", itid))
					}
					out = append(out, s(it.Invocations)...)
				}
			}
		}
	}

	return out
}

func (e *Execution) AddInvocations(is InvocationsFlat) error {
	if e == nil {
		return fmt.Errorf("Execution e is nil")
	}

	if !e.Benchmark.Equals(is.Benchmark) {
		return fmt.Errorf("Execution belongs to %v, not to %v", e.Benchmark, is.Benchmark)
	}

	// create Execution from InvocationsFlat
	ne := NewExecutionFromInvocationsFlat(is)
	err := e.Merge(ne)
	if err != nil {
		return err
	}

	e.addLen(ne.len)

	return nil
}

func (e *Execution) Merge(other *Execution) error {
	if !e.Benchmark.Equals(other.Benchmark) {
		return fmt.Errorf("Benchmarks not the same: this %+v, other %+v", e.Benchmark, other.Benchmark)
	}

	for _, oiid := range other.InstanceIDs {
		oi, ok := other.Instances[oiid]
		if !ok {
			panic(fmt.Sprintf("Invalid state: InstanceIDs and Instances out of sync for %s", oiid))
		}
		i, ok := e.Instances[oiid]
		if !ok {
			e.InstanceIDs = append(e.InstanceIDs, oiid)
			e.Instances[oiid] = oi
		} else {
			err := i.Merge(oi)
			if err != nil {
				return err
			}
		}
	}

	// update length
	other.lenLock.RLock()
	defer other.lenLock.RUnlock()
	e.addLen(other.len)

	return nil
}

type Executions []*Execution

func (e Executions) Len() int {
	return len(e)
}

func (e Executions) Less(i, j int) bool {
	// return true iff i is smaller or equal to j
	return e[j].Benchmark.Compare(e[i].Benchmark) == -1
}

func (e Executions) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
