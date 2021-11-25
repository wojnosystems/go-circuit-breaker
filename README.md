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
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	breaker := twoStateCircuit.New(twoStateCircuit.Opts{
		FailureThreshold: 2,
		OpenDuration:     30 * time.Second,
	})
	s := &SDK{
		baseUrl: "https://example.com/api",
		httpClient: circuitHTTP.New(breaker, http.DefaultClient),
	}
	_, _ = s.MakeThing("1")
	_, _ = s.MakeThing("2")
	_, _ = s.MakeThing("3")
	_, _ = s.MakeThing("4")
}

type SDK struct {
	httpClient *circuitHTTP.Client
	baseUrl string
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
	resp, err := s.httpClient.Post(s.baseUrl + "/thing", "application/json", bytes.NewBuffer(serializedThing))
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

## Example: Logging open and close state transitions

This example uses the built-in channel to log when state transitions occur:

```go
package main

import (
	"github.com/wojnosystems/go-circuit-breaker/circuitHTTP"
	"github.com/wojnosystems/go-circuit-breaker/twoStateCircuit"
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
		FailureThreshold: 2,
		OpenDuration:     30 * time.Second,
		OnStateChange:    stateTransition,
	})
	client := circuitHTTP.New(breaker, http.DefaultClient)

	_, _ = client.Get("https://example.com/api/things/1")
	// do more things with the breaker
}
```
