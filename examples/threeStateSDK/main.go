package main

import (
	"github.com/wojnosystems/go-circuit-breaker/circuitHTTP"
	"github.com/wojnosystems/go-circuit-breaker/threeStateCircuit"
	"github.com/wojnosystems/go-rate-limit/rateLimit"
	"net/http"
	"time"
)

func main() {
	breaker := threeStateCircuit.New(threeStateCircuit.OptsWithTokenBucketTripDecider(
		threeStateCircuit.Opts{
			// Breaker will wait 30 seconds in the Open State before transitioning to Half-Open
			OpenDuration: 30 * time.Second,
			// When in Half-Open, will sample up to 50% of requests, becoming more likely
			// over the course of 60 seconds. When the breaker first enters Half-Open,
			// the chance of being sampled is 0, slowly increasing to 50% once 60 seconds have passed
			// Feel free to swap this out with whatever you need.
			HalfOpenSampler: threeStateCircuit.NewLinearScalingSamplerWithStandardRandom(
				60*time.Second,
				.5,
			),
			NumberOfSuccessesInHalfOpenToClose: 5,
		},
		tokenBucketOptions,
	))

	httpClient := circuitHTTP.New(breaker, http.DefaultClient)

	httpClient.Get("https://www.example.com/broken")
	httpClient.Get("https://www.example.com/broken")
	httpClient.Get("https://www.example.com/broken")
	httpClient.Get("https://www.example.com/broken")
}

var tokenBucketOptions = rateLimit.TokenBucketOpts{
	// We only allow up to 2 errors per second
	Capacity:             2,
	TokensAddedPerSecond: 2,
	// Prime the breaker with 2 errors allowed at start, you could set this to 0 and force the breaker
	// to "charge" before use
	InitialTokens: 2,
}
