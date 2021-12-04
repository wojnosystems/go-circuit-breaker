package threeStateCircuit

import (
	. "github.com/onsi/gomega"
	"math"
	"math/rand"
	"testing"
	"time"
)

type randSrcAlwaysZero struct {
}

func (r *randSrcAlwaysZero) Int63() int64 {
	return 0
}

func (r *randSrcAlwaysZero) Seed(_ int64) {
}

type randSrcAlwaysOne struct {
}

func (r *randSrcAlwaysOne) Int63() int64 {
	return math.MaxInt64 - 512
}

func (r *randSrcAlwaysOne) Seed(_ int64) {
}

func TestNewLinearScalingSampler(t *testing.T) {
	cases := map[string]struct {
		inputDuration   time.Duration
		inputMaxPercent float64
		inputRandSource rand.Source
		timeInHalfOpen  time.Duration
		expected        bool
	}{
		"when no time spent, then do not sample": {
			inputDuration:   30 * time.Second,
			inputMaxPercent: .25,
			inputRandSource: &randSrcAlwaysZero{},
			timeInHalfOpen:  0,
			expected:        false,
		},
		"all time spent, then sample": {
			inputDuration:   30 * time.Second,
			inputMaxPercent: 1.0,
			inputRandSource: &randSrcAlwaysOne{},
			timeInHalfOpen:  30 * time.Second,
			expected:        true,
		},
	}
	for caseName, dt := range cases {
		t.Run(caseName, func(t *testing.T) {
			g := NewWithT(t)
			subject := NewLinearScalingSampler(dt.inputDuration, dt.inputMaxPercent, dt.inputRandSource)
			actual := subject.ShouldSample(dt.timeInHalfOpen)
			g.Expect(actual).Should(Equal(dt.expected))
		})
	}
}
