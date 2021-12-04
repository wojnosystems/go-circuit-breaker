package tripping

import (
	"errors"
	. "github.com/onsi/gomega"
	"testing"
)

var wrappedError = errors.New("wrapped")

func TestError_Error(t *testing.T) {
	cases := map[string]struct {
		builder      func() *Error
		expected     string
		expectedCost uint64
	}{
		"unit constructor": {
			builder: func() *Error {
				return New(wrappedError)
			},
			expected:     wrappedError.Error(),
			expectedCost: 1,
		},
		"custom cost constructor": {
			builder: func() *Error {
				return NewWithCost(wrappedError, 5)
			},
			expected:     wrappedError.Error(),
			expectedCost: 5,
		},
	}
	for caseName, dt := range cases {
		t.Run(caseName, func(t *testing.T) {
			g := NewWithT(t)
			actual := dt.builder()
			g.Expect(actual.Error()).Should(Equal(dt.expected))
			g.Expect(actual.Cost).Should(Equal(dt.expectedCost))
		})
	}
}

func TestIsTripping(t *testing.T) {
	cases := map[string]struct {
		input    error
		expected bool
	}{
		"nil": {},
		"non-tripping": {
			input: errors.New("not tripping"),
		},
		"tripping": {
			input:    New(wrappedError),
			expected: true,
		},
	}
	for caseName, dt := range cases {
		t.Run(caseName, func(t *testing.T) {
			g := NewWithT(t)
			actual := IsTripping(dt.input)
			g.Expect(actual).Should(Equal(dt.expected))
		})
	}
}
