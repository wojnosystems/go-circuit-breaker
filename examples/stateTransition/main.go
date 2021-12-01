package main

import (
	"github.com/wojnosystems/go-circuit-breaker/circuitHTTP"
	"github.com/wojnosystems/go-circuit-breaker/twoStateCircuit"
	"github.com/wojnosystems/go-rate-limit/rateLimit"
	"log"
	"net/http"
	"time"
)

func main() {
	stateTransition := make(chan twoStateCircuit.State, 10)
	go func() {
		for {
			newState, ok := <-stateTransition
			if !ok {
				return
			}
			log.Println("state is now:", newState.String())
		}
	}()

	breaker := twoStateCircuit.New(twoStateCircuit.Opts{
		FailureLimiter: rateLimit.NewTokenBucket(rateLimit.TokenBucketOpts{
			Capacity:             2,
			TokensAddedPerSecond: 2,
			InitialTokens:        2,
		}),
		OpenDuration:  30 * time.Second,
		OnStateChange: stateTransition,
	})
	client := circuitHTTP.New(breaker, http.DefaultClient)

	_, _ = client.Get("https://example.com/api/things/1")
	// do more things with the breaker
}
