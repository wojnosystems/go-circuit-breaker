package main

import (
	"github.com/wojnosystems/go-circuit-breaker/circuitHTTP"
	"github.com/wojnosystems/go-circuit-breaker/twoStateCircuit"
	"github.com/wojnosystems/go-circuit-breaker/twoStateCircuit/state"
	"github.com/wojnosystems/go-rate-limit/rateLimit"
	"log"
	"net/http"
	"time"
)

func main() {
	// stateTransition is a channel that will receive events when the state changes.
	stateTransition := make(chan state.State, 10)
	go func() {
		for {
			newState, ok := <-stateTransition
			if !ok {
				return
			}
			log.Println("state is now:", newState.String())
		}
	}()

	breaker := twoStateCircuit.New(twoStateCircuit.OptsWithTokenBucketTripDecider(
		twoStateCircuit.Opts{
			OpenDuration: 30 * time.Second,
			// OnStateChange is assigned the channel from above, this is how we tell the breaker to send events
			OnStateChange: stateTransition,
		},
		tokenBucketOptions,
	))
	client := circuitHTTP.New(breaker, http.DefaultClient)

	_, _ = client.Get("https://example.com/api/things/1")
	// do more things with the breaker
}

var tokenBucketOptions = rateLimit.TokenBucketOpts{
	// We only allow up to 2 errors per second
	Capacity:             2,
	TokensAddedPerSecond: 2,
	// Prime the breaker with 2 errors allowed at start, you could set this to 0 and force the breaker
	// to "charge" before use
	InitialTokens: 2,
}
