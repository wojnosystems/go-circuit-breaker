package threeStateCircuit

import (
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

func TestShouldSample_ShouldSample(t *testing.T) {
	cases := map[string]struct {
		build    func() ShouldSample
		input    time.Duration
		expected bool
	}{
		"default": {
			build: func() (x ShouldSample) {
				return
			},
			input:    1 * time.Second,
			expected: true,
		},
		"overridden": {
			build: func() ShouldSample {
				return func(_ time.Duration) bool {
					return false
				}
			},
			input: 1 * time.Second,
		},
	}

	for caseName, dt := range cases {
		t.Run(caseName, func(t *testing.T) {
			g := NewWithT(t)
			actual := dt.build().ShouldSample(dt.input)
			g.Expect(actual).Should(Equal(dt.expected))
		})
	}
}
