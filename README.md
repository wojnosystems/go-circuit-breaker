# Overview

Example library for [circuit breaking](https://martinfowler.com/bliki/CircuitBreaker.html) in GoLang. Written to support a blog post on [https://www.wojno.com](https://www.wojno.com).

Use this library in your SDK's to prevent overwhelming your backend servers during an outage or time of extreme traffic.

# How to Use

## Install

`go get -u github.com/wojnosystems/go-circuit-breaker`

## Example: SDK Client

A mock SDK using the breaker:

```go
package main

import (
	"bytes"
	"encoding/json"
	"github.com/wojnosystems/go-circuit-breaker/circuitHTTP"
	"github.com/wojnosystems/go-circuit-breaker/twoStateCircuit"
	"github.com/wojnosystems/go-rate-limit/rateLimit"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	breaker := twoStateCircuit.New(twoStateCircuit.OptsWithTokenBucketTripDecider(
		// Create a two-state breaker that will stay in the open state for 30 seconds
		twoStateCircuit.Opts{
			OpenDuration: 30 * time.Second,
		},
		// This breaker will use a token bucket to track error rates. When exceeded, the breaker will trip
		rateLimit.TokenBucketOpts{
			// We only allow up to 2 errors per second
			Capacity:             2,
			TokensAddedPerSecond: 2,
			// Prime the breaker with 2 errors allowed at start, you could set this to 0 and force the breaker
			// to "charge" before use
			InitialTokens: 2,
		},
	))
	s := &SDK{
		baseUrl: "https://example.com/api",
		// Install the breaker we created above into the http client. Any failing requests will interact with the breaker
		httpClient: circuitHTTP.New(breaker, http.DefaultClient),
	}
	_, _ = s.MakeThing("1")
	_, _ = s.MakeThing("2")
	_, _ = s.MakeThing("3")
	_, _ = s.MakeThing("4")
}

type SDK struct {
	httpClient *circuitHTTP.Client
	baseUrl    string
}

type Thing struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

func (s *SDK) MakeThing(name string) (thingId uint64, err error) {
	t := Thing{
		Name: name,
	}
	serializedThing, err := json.Marshal(t)
	if err != nil {
		return
	}
	resp, err := s.httpClient.Post(s.baseUrl+"/thing", "application/json", bytes.NewBuffer(serializedThing))
	if err != nil {
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &t)
	if err != nil {
		return
	}
	thingId = t.Id
	return
}
```

In the above example, we created a new circuit breaker, then wrapped it in an `http.Client`, then used it in our mocked SDK. This allows all endpoints using http to take advantage of a circuit breaker without actually changing any code that uses the real http.Client.

## Example: Logging open and close state transitions

This example uses the built-in channel to log when state transitions occur:

```go
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
	// stateTransition is a channel that will receive events when the state changes.
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
```

Each time the circuit breaker change from open to closed or closed to open, it will append the new state to the channel `stateTransition`.

Be very careful not to let this channel fill up or all of your requests will block. It's also important not to close this channel, otherwise the circuit breaker will attempt to use a closed channel and panic.

## Three-State Breaker: Closed, Open, Half-Open

Here's an example of using a three-state breaker. These are better than the simple two-state breakers as they allow for more rapid recovery while still avoiding overwhelming the backing service. In this example, we create a three-state breaker that uses a Token Bucket to determine if there are too many errors in the closed state, it will transition to the open state for the OpenStateDuration. After this time has passed from tripping, The token bucket will sample requests. Any tripping errors during this time will revert the breaker to the open state. If 5 requests succeed in a row, then the breaker will transition back to the Closed state, and all requests will be executed.


```go
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
```

# Future work

The next steps area to combine the power of circuit breakers and retry logic. One could easily wrap a circuit breaker in a retry block so that requests are more robust without significantly adding burden to the backend.
