package bench

import "math"

type ExecutionTransformer interface {
	// transform transforms an execution `e` into a new execution (should return a copy of the execution)
	transform(e *Execution) *Execution
}

var _ ExecutionTransformer = NamedExecutionTransformer{}

type NamedExecutionTransformer struct {
	ExecutionTransformer
	Name string
}

var _ ExecutionTransformer = ExecutionTransformerFunc(func(e *Execution) *Execution { return e })

type ExecutionTransformerFunc func(*Execution) *Execution

func (f ExecutionTransformerFunc) transform(e *Execution) *Execution {
	return f(e)
}

func IdentityExecutionTransformerFunc(e *Execution) *Execution {
	return e
}

func ConstantFactorExecutionTransformerFunc(factor float64, roundingPrecision int) ExecutionTransformerFunc {
	roundingFactor := math.Pow(10, float64(roundingPrecision))
	return func(e *Execution) *Execution {
		ne := e.Copy()

		for _, instance := range ne.Instances {
			for _, trial := range instance.Trials {
				for _, fork := range trial.Forks {
					for _, iteration := range fork.Iterations {
						newInvocations := make([]Invocations, len(iteration.Invocations))
						for i, invocations := range iteration.Invocations {
							newInvocations[i] = Invocations{
								Count: invocations.Count,
								Value: math.Round(invocations.Value*factor*roundingFactor) / roundingFactor,
							}
						}
						iteration.Invocations = newInvocations
					}
				}
			}
		}

		return ne
	}
}
