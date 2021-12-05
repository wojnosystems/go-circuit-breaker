package twoStateCircuit

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/wojnosystems/go-circuit-breaker/tripping"
	"github.com/wojnosystems/go-circuit-breaker/twoStateCircuit/state"
	"github.com/wojnosystems/go-time-factory/timeFactory"
	"time"
)

var trippingError = tripping.New(errors.New("tripping error"))

var _ = Describe("Breaker.Use", func() {
	When("used without errors", func() {
		var (
			subject     *Breaker
			stateChange chan state.State
		)
		BeforeEach(func() {
			stateChange = make(chan state.State, 10)
			subject = New(Opts{
				TripDecider:   neverTrips,
				OpenDuration:  1 * time.Hour,
				OnStateChange: stateChange,
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
	When("error threshold not met", func() {
		var (
			subject     *Breaker
			stateChange chan state.State
			options     Opts
			startTime   time.Time
		)
		BeforeEach(func() {
			startTime = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
			stateChange = make(chan state.State, 10)
			options = Opts{
				TripDecider:   neverTrips,
				OpenDuration:  50 * time.Millisecond,
				OnStateChange: stateChange,
				nowFactory: func() time.Time {
					return startTime
				},
			}
			subject = New(options)

			_ = subject.Use(func() error {
				return trippingError
			})
		})
		It("calls the callback", func() {
			called := false
			_ = subject.Use(func() error {
				called = true
				return nil
			})
			Expect(called).Should(BeTrue())
		})
		It("does not transition", func() {
			Eventually(stateChange).ShouldNot(Receive())
		})
	})
	When("error threshold met", func() {
		var (
			subject     *Breaker
			stateChange chan state.State
			options     Opts
			startTime   time.Time
		)
		BeforeEach(func() {
			startTime = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
			stateChange = make(chan state.State, 10)
			options = Opts{
				OpenDuration:  50 * time.Millisecond,
				OnStateChange: stateChange,
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
				Expect(err).Should(Equal(trippingError.Err))
			})
		})
		It("notifies state is open", func() {
			Expect(stateChange).Should(Receive(Equal(state.Open)))
		})
		When("open state time expires", func() {
			BeforeEach(func() {
				Expect(stateChange).Should(Receive(Equal(state.Open)))
				subject.state = state.Open
				subject.mutableState.openExpiresAt = time.Now().Add(-1 * time.Second)
				subject.opts.nowFactory = timeFactory.ReturnTimes(
					time.Now(),
					time.Now(),
				)
			})
			It("notifies state is closed", func() {
				_ = subject.Use(func() error {
					return nil
				})
				Expect(stateChange).Should(Receive(Equal(state.Closed)))
			})
		})
	})
})
