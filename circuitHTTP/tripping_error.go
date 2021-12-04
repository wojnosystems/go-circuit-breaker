package circuitHTTP

import (
	"errors"
	"github.com/wojnosystems/go-circuit-breaker/tripping"
	"net/http"
)

// ConvertToTrippingErrIfShould converts the http response and error, if any, into a tripping error
// If you don't want to trip on this error, simply return nil or the original error.
// If you want this error to contribute to tripping the breaker, wrap the error in a tripping.New or tripping.NewWithCost to mark
// the error as able to trip the breaker.
type ConvertToTrippingErrIfShould func(resp *http.Response, err error) error

// ConvertToTrippingErrIfShould functor that allows the default to be called without checking for nil in the Client
func (s ConvertToTrippingErrIfShould) ConvertToTrippingErrIfShould(resp *http.Response, err error) error {
	if s != nil {
		return s(resp, err)
	}
	return defaultConvertToTrippingErrIfShould(resp, err)
}

type Breaker interface {
	Use(callback func() error) error
}

func defaultConvertToTrippingErrIfShould(resp *http.Response, err error) error {
	// all errors trip the breaker
	if err != nil {
		return tripping.New(err)
	}
	// Some status codes also trip the breaker, even if there was no error
	switch resp.StatusCode {
	// TODO: custom errors for each condition
	case http.StatusBadGateway, // usually coincides with the backend being down
		http.StatusInternalServerError, // some services will throw this when overwhelmed, too
		http.StatusRequestTimeout,      // servers usually are programmed to return this when their own req timeout expires
		http.StatusServiceUnavailable,  // usually coincides with the backend being down
		http.StatusTooManyRequests:     // service has a rate limiter telling us to slow down
		return tripping.New(errors.New("upstream service is down or is rateLimit-limiting"))
	default:
		return err
	}
}
