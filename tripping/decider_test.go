package tripping

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestDecider_ShouldTrip(t *testing.T) {
	cases := map[string]struct {
		input    Decider
		expected bool
	}{
		"not assigned": {
			expected: true,
		},
		"assigned": {
			input: func(_ *Error) (shouldTrip bool) {
				return false
			},
		},
	}
	for caseName, dt := range cases {
		t.Run(caseName, func(t *testing.T) {
			g := NewGomegaWithT(t)
			actual := dt.input.ShouldTrip(New(wrappedError))
			g.Expect(actual).Should(Equal(dt.expected))
		})
	}
}
