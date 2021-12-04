package threeStateCircuit

import (
	"math/rand"
	"sync"
	"time"
)

// NewLinearScalingSampler will allow the circuit breaker to test a percentage of requests while in the Half Open State.
// The chances of the request being tested increases linearly with time.
// The longer it's in the half-open state, the more likely that the request will be attempted, up to the maximumChance
// For example, if you set:
//   scaleOverDuration: 30 * time.Second
//   maximumChance: 0.25
//
// If the breaker has been in the half-open state for 10 seconds, there is a 25% * 33% or 8.25% chance that
// the circuit breaker will attempt the request. Approximately every 1 in 12 usage attempts will actually send a request.
// After 30 seconds, every request has a 25% chance or roughly 1 out of every 4 requests of being attempted.
// The randomSource allows you to specify a custom source of randomness, in case you want to seed it or use something
// different. Usually, the pseudo-random number generator provided by math.Rand will suffice.
// The randomSource does not need to be thread-safe, this method will ensure that it's not used concurrently,
// assuming no other threads are also using this same random source.
//
// Example:
// sampler := NewLinearScalingSampler(30 * time.Second, 0.25, rand.NewSource(time.Now().UnixNano()))
// sampler.ShouldSample(timeInHalfOpen) -> true/false
func NewLinearScalingSampler(scaleOverDuration time.Duration, maximumChance float64, randomSource rand.Source) ShouldSample {
	randSource := rand.New(randomSource)
	var mu sync.Mutex
	return func(timeInHalfOpen time.Duration) (shouldSample bool) {
		mu.Lock()
		randomValue := randSource.Float64()
		mu.Unlock()
		if timeInHalfOpen >= scaleOverDuration {
			return randomValue < maximumChance
		}
		chancePercent := float64(timeInHalfOpen) / float64(scaleOverDuration) * maximumChance
		return randomValue < chancePercent
	}
}
