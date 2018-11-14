package bench

import (
	"fmt"
	"sort"
)

type Chan chan ExecutionValue

type ExecutionValue struct {
	Type ExecutionType
	Exec *Execution
	Err  error
}

type ExecutionType int

const (
	ExecNext ExecutionType = iota
	ExecError
	ExecStart
	ExecEnd
)

func MergeChans(cs ...Chan) Chan {
	out := make(Chan)

	go func() {
		defer close(out)

		active := len(cs)
		crs := make(Executions, 0, len(cs))

		out <- ExecutionValue{
			Type: ExecStart,
		}

		for active != 0 {
		ChanLoop:
			for _, c := range cs {
				ev, ok := <-c
				if !ok {
					// remove channel from array of channels
					continue
				}

				switch ev.Type {
				case ExecStart:
					continue ChanLoop
				case ExecEnd:
					active--
					continue ChanLoop
				case ExecError:
					out <- ev
				case ExecNext:
					crs = append(crs, ev.Exec)
				}
			}

			mergeExecutionValues(crs, out)

			// reset crs
			crs = make(Executions, 0, len(cs))
		}

		out <- ExecutionValue{
			Type: ExecEnd,
		}
	}()

	return out
}

func mergeExecutionValues(evs Executions, out Chan) {
	if len(evs) == 0 {
		return
	}
	// sort evs
	sort.Sort(evs)

	// merge the ones that are equal, starting from the front
	var prev *Execution
	for _, ev := range evs {
		if prev != nil {
			if prev.Benchmark.Equals(ev.Benchmark) {
				// perform merge
				err := prev.Merge(ev)
				if err != nil {
					panic(fmt.Sprintf("Could not merge Executions: %v", err))
				}
				continue
			} else {
				// send unmerged
				out <- ExecutionValue{
					Type: ExecNext,
					Exec: prev,
				}
			}
		}
		prev = ev
	}

	out <- ExecutionValue{
		Type: ExecNext,
		Exec: prev,
	}
}
