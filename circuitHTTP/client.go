package circuitHTTP

import (
	"errors"
	"github.com/wojnosystems/go-circuit-breaker/tripping"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ReturnTrippingIfShould func(resp *http.Response, err error) error

type Breaker interface {
	Use(callback func() error) error
}

func defaultTrippedDecision(resp *http.Response, err error) error {
	// all errors trip the breaker
	if err != nil {
		return tripping.New(err)
	}
	switch resp.StatusCode {
	// TODO: custom errors for each condition
	case http.StatusBadGateway, http.StatusInternalServerError, http.StatusRequestTimeout, http.StatusServiceUnavailable, http.StatusTooManyRequests:
		return tripping.New(errors.New("upstream service is down or is rateLimit-limiting"))
	default:
		return nil
	}
}

type Client struct {
	*http.Client
	breaker     Breaker
	tripDecider ReturnTrippingIfShould
}

func New(breaker Breaker, client *http.Client) *Client {
	return NewWithTripDecider(breaker, client, defaultTrippedDecision)
}

func NewWithTripDecider(breaker Breaker, client *http.Client, tripDecider ReturnTrippingIfShould) *Client {
	return &Client{
		Client:      client,
		breaker:     breaker,
		tripDecider: tripDecider,
	}
}

func (c *Client) Do(req *http.Request) (resp *http.Response, err error) {
	err = c.breaker.Use(func() error {
		resp, err = c.Client.Do(req)
		return c.decideToTrip(resp, err)
	})
	return
}

func (c *Client) decideToTrip(resp *http.Response, err error) error {
	if c.tripDecider != nil {
		return c.tripDecider(resp, err)
	}
	return defaultTrippedDecision(resp, err)
}

func (c *Client) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Head(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

func (c *Client) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}
