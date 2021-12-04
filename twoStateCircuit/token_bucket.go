package twoStateCircuit

import (
	"github.com/wojnosystems/go-circuit-breaker/tripping"
	"github.com/wojnosystems/go-rate-limit/rateLimit"
)

// OptsWithTokenBucketTripDecider creates a new breaker backed by a token bucket limiter
func OptsWithTokenBucketTripDecider(breakerOpts Opts, tokenBucketOpts rateLimit.TokenBucketOpts) Opts {
	tokenBucket := rateLimit.NewTokenBucket(tokenBucketOpts)
	breakerOpts.TripDecider = func(trippingErr *tripping.Error) (shouldTrip bool) {
		return !tokenBucket.Allowed(trippingErr.Cost)
	}
	return breakerOpts
}
