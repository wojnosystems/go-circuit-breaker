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
