package bench_test

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/chrstphlbr/pa/pkg/bench"
)

func TestIdentityExecutionTransformerFunc(t *testing.T) {
	e := complexExecution(t)
	eid := bench.IdentityExecutionTransformerFunc(e)

	if e != eid {
		t.Fatalf("identity execution not the same")
	}

	if e.Benchmark != eid.Benchmark {
		t.Fatalf("benchmark not the same")
	}

	if !e.Benchmark.Equals(eid.Benchmark) {
		t.Fatalf("benchmark not equal")
	}

	equalInstances(t, e, eid, false)
}

func TestConstantFactorExecutionTransformerFunc(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	e := complexExecution(t)
	factor := rnd.Float64()
	roundingPrecision := 5
	roundingFactor := math.Pow(10, float64(roundingPrecision))
	transformer := bench.ConstantFactorExecutionTransformerFunc(factor, roundingPrecision)

	te := transformer(e)

	for instanceID, instance := range te.Instances {
		for trialID, trial := range instance.Trials {
			for forkID, fork := range trial.Forks {
				for iterationID, iteration := range fork.Iterations {
					for invocationPos, invocation := range iteration.Invocations {
						invocationPrevious := e.Instances[instanceID].Trials[trialID].Forks[forkID].Iterations[iterationID].Invocations[invocationPos]

						if invocation.Count != invocationPrevious.Count {
							t.Fatalf("Invocation.Count has been altered, should be %d, was %d", invocationPrevious.Count, invocation.Count)
						}

						expectedValue := math.Round(invocationPrevious.Value*factor*roundingFactor) / roundingFactor
						if invocation.Value != expectedValue {
							t.Fatalf("Invocation.Value was not transformed correctly: should be %f * %f = %f, was %f", invocationPrevious.Value, factor, expectedValue, invocation.Value)
						}

					}
				}
			}
		}
	}
}
