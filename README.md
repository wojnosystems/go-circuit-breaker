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
	breaker := twoStateCircuit.New(twoStateCircuit.Opts{
		FailureLimiter: rateLimit.NewTokenBucket(rateLimit.TokenBucketOpts{
			Capacity:             2,
			TokensAddedPerSecond: 2,
			InitialTokens:        2,
		}),
		OpenDuration:   30 * time.Second,
	})
	s := &SDK{
		baseUrl:    "https://example.com/api",
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
	stateTransition := make(chan twoStateCircuit.State, 10)
	go func() {
		for {
			newState, ok := <-stateTransition
			if !ok {
				return
			}
			log.Println("state is now:", newState)
		}
	}()

	breaker := twoStateCircuit.New(twoStateCircuit.Opts{
		FailureLimiter: rateLimit.NewTokenBucket(rateLimit.TokenBucketOpts{
			Capacity:             2,
			TokensAddedPerSecond: 2,
			InitialTokens:        2,
		}),
		OpenDuration:     30 * time.Second,
		OnStateChange:    stateTransition,
	})
	client := circuitHTTP.New(breaker, http.DefaultClient)

	_, _ = client.Get("https://example.com/api/things/1")
	// do more things with the breaker
}
```

Each time the circuit breaker change from open to closed or closed to open, it will append the new state to the channel `stateTransition`.

Be very careful not to let this channel fill up or all of your requests will block. It's also important not to close this channel, otherwise the circuit breaker will attempt to use a closed channel and panic.

# Future work

The next steps area to combine the power of circuit breakers and retry logic. One could easily wrap a circuit breaker in a retry block so that requests 
