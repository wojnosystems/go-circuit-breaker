package twoStateCircuit

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/wojnosystems/go-circuit-breaker/circuitTripping"
	"time"
)

var trippingError = circuitTripping.New(errors.New("tripping error"))

var _ = Describe("Breaker", func() {
	When("used without errors", func() {
		var (
			subject     *Breaker
			stateChange chan State
		)
		BeforeEach(func() {
			stateChange = make(chan State, 10)
			subject = New(Opts{
				FailureLimiter: &tokenBucketAlwaysSucceeds{},
				OpenDuration:   1 * time.Hour,
				OnStateChange:  stateChange,
			})

			for i := 0; i < 10; i++ {
				_ = subject.Use(func() error {
					return nil
				})
			}
		})
		It("continues working", func() {
			err := subject.Use(func() error {
				return nil
			})
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("does not notify state change", func() {
			Expect(stateChange).ShouldNot(Receive())
		})
	})

	When("error threshold met", func() {
		var (
			subject     *Breaker
			stateChange chan State
			options     Opts
			startTime   time.Time
		)
		BeforeEach(func() {
			startTime = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
			stateChange = make(chan State, 10)
			options = Opts{
				FailureLimiter: &tokenBucketAlwaysFails{},
				OpenDuration:   50 * time.Millisecond,
				OnStateChange:  stateChange,
				nowFactory: func() time.Time {
					return startTime
				},
			}
			subject = New(options)

			_ = subject.Use(func() error {
				return trippingError
			})
		})
		When("used while open", func() {
			var (
				err error
			)
			BeforeEach(func() {
				err = subject.Use(func() error {
					return nil
				})
			})
			It("returns the last error", func() {
				Expect(err).Should(Equal(trippingError.Unwrap()))
			})
		})
		It("notifies state is open", func() {
			Expect(stateChange).Should(Receive(Equal(StateOpen)))
		})
		When("open state time expires", func() {
			BeforeEach(func() {
				subject.opts.nowFactory = func() time.Time {
					return startTime.Add(subject.opts.OpenDuration)
				}
				// receive state open notice
				Expect(stateChange).Should(Receive(Equal(StateOpen)))
			})
			It("notifies state is open", func() {
				_ = subject.Use(func() error {
					return nil
				})
				Expect(stateChange).Should(Receive(Equal(StateClosed)))
			})
		})
	})
})

type tokenBucketAlwaysFails struct{}

func (b *tokenBucketAlwaysFails) Allowed(_ uint64) bool {
	return false
}

type tokenBucketAlwaysSucceeds struct{}

func (s *tokenBucketAlwaysSucceeds) Allowed(_ uint64) bool {
	return true
}
