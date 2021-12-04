package main

import (
	"bytes"
	"encoding/json"
	"github.com/wojnosystems/go-circuit-breaker/circuitHTTP"
	"github.com/wojnosystems/go-circuit-breaker/tripping"
	"github.com/wojnosystems/go-circuit-breaker/twoStateCircuit"
	"github.com/wojnosystems/go-rate-limit/rateLimit"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	tokenBucket := rateLimit.NewTokenBucket(rateLimit.TokenBucketOpts{
		Capacity:             2,
		TokensAddedPerSecond: 2,
		InitialTokens:        2,
	})
	breaker := twoStateCircuit.New(twoStateCircuit.Opts{
		TripDecider: func(trippingErr *tripping.Error) (shouldTrip bool) {
			return !tokenBucket.Allowed(trippingErr.Cost)
		},
		OpenDuration: 30 * time.Second,
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
