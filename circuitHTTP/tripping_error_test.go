package circuitHTTP

import (
	"errors"
	. "github.com/onsi/gomega"
	"github.com/wojnosystems/go-circuit-breaker/tripping"
	"net/http"
	"testing"
)

func Test_defaultConvertToTrippingErrIfShould(t *testing.T) {
	cases := map[string]struct {
		inputErr         error
		inputResp        *http.Response
		expectedTripping bool
	}{
		"error trips": {
			inputErr:         errors.New("some error"),
			expectedTripping: true,
		},
		"ok does not trip": {
			inputResp: &http.Response{
				StatusCode: http.StatusOK,
			},
		},
		"rate limit status trips": {
			inputResp: &http.Response{
				StatusCode: http.StatusTooManyRequests,
			},
			expectedTripping: true,
		},
	}
	for caseName, dt := range cases {
		t.Run(caseName, func(t *testing.T) {
			g := NewWithT(t)
			actual := tripping.IsTripping(defaultConvertToTrippingErrIfShould(dt.inputResp, dt.inputErr))
			g.Expect(actual).Should(Equal(dt.expectedTripping))
		})
	}
}

func TestConvertToTrippingErrIfShould_ConvertToTrippingErrIfShould(t *testing.T) {
	defaultError := errors.New("default")
	overriddenError := errors.New("overridden")
	cases := map[string]struct {
		build       func() ConvertToTrippingErrIfShould
		inputErr    error
		expectedErr error
	}{
		"default": {
			build: func() (x ConvertToTrippingErrIfShould) {
				return
			},
			inputErr:    defaultError,
			expectedErr: tripping.New(defaultError),
		},
		"overridden": {
			build: func() ConvertToTrippingErrIfShould {
				return func(_ *http.Response, _ error) error {
					return overriddenError
				}
			},
			inputErr:    nil,
			expectedErr: overriddenError,
		},
	}

	for caseName, dt := range cases {
		t.Run(caseName, func(t *testing.T) {
			g := NewWithT(t)
			actual := dt.build().ConvertToTrippingErrIfShould(&http.Response{}, dt.inputErr)
			g.Expect(actual).Should(Equal(dt.expectedErr))
		})
	}
}
