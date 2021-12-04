package circuitHTTP

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client is a http.Client with a circuit breaker inside. Every request that fails is counted in the breaker
// Use New or NewWithTripDecider instead of using this struct as breaker and tripDecider require initialization
type Client struct {
	*http.Client
	breaker     Breaker
	tripDecider ConvertToTrippingErrIfShould
}

// New creates a new http.Client with a breaker inside
// by default, the breaker can trip when the client receives timeouts or http statuses that usually indicate
// an outage or rate limit.
func New(breaker Breaker, client *http.Client) *Client {
	return NewWithTripDecider(breaker, client, defaultConvertToTrippingErrIfShould)
}

// NewWithTripDecider is like New, but allows you to customize which http statuses or errors trip the breaker
func NewWithTripDecider(breaker Breaker, client *http.Client, tripDecider ConvertToTrippingErrIfShould) *Client {
	return &Client{
		Client:      client,
		breaker:     breaker,
		tripDecider: tripDecider,
	}
}

func (c *Client) Do(req *http.Request) (resp *http.Response, err error) {
	err = c.breaker.Use(func() error {
		resp, err = c.Client.Do(req)
		return c.tripDecider.ConvertToTrippingErrIfShould(resp, err)
	})
	return
}

func (c *Client) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Head(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

func (c *Client) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}
