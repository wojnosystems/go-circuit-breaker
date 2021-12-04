package threeStateCircuit

import "time"

type ShouldSample func(timeInHalfOpen time.Duration) (shouldSample bool)

// ShouldSample returns true if the attempt should be attempted while in the half-open state
// Use this to set a percentage or sliding scale based on time for when to attempt a request while in the half-open state
func (s ShouldSample) ShouldSample(timeInHalfOpen time.Duration) bool {
	if s != nil {
		return s(timeInHalfOpen)
	}
	return true
}
