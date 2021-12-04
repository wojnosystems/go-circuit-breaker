package threeStateCircuit

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/wojnosystems/go-circuit-breaker/tripping"
	"time"
)

var trippingError = tripping.New(errors.New("force trip"))

func samplerAlwaysSamples(timeInHalfOpen time.Duration) bool {
	return true
}
func samplerNeverSamples(timeInHalfOpen time.Duration) bool {
	return false
}

var _ = Describe("Breaker.Use", func() {
	var (
		stateChange chan State
	)
	BeforeEach(func() {
		stateChange = make(chan State, 10)
	})
	When("closed", func() {
		var (
			breaker *Breaker
		)
		BeforeEach(func() {
			breaker = New(Opts{
				TripDecider: func(_ *tripping.Error) (shouldTrip bool) {
					return true
				},
				OpenDuration:                       0,
				OnStateChange:                      stateChange,
				HalfOpenSampler:                    nil,
				NumberOfSuccessesInHalfOpenToClose: 0,
			})
		})
		When("tripping", func() {
			When("error threshold not exceeded", func() {
				BeforeEach(func() {
					breaker.opts.TripDecider = func(_ *tripping.Error) (shouldTrip bool) {
						return false
					}
				})
				It("does not transition", func() {
					_ = breaker.Use(func() error {
						return trippingError
					})
					Eventually(stateChange).ShouldNot(Receive())
				})
			})

			When("error threshold exceeded", func() {
				It("transitions to open", func() {
					_ = breaker.Use(func() error {
						return trippingError
					})
					Expect(stateChange).Should(Receive(Equal(StateOpen)))
				})
			})

			When("open transition occurred while testing", func() {
				It("does not transition", func() {
					_ = breaker.Use(func() error {
						breaker.state = StateOpen
						return trippingError
					})
					Eventually(stateChange).ShouldNot(Receive())
				})
			})
		})
	})
	When("open", func() {
		var (
			breaker *Breaker
		)
		BeforeEach(func() {
			breaker = New(Opts{
				TripDecider: func(_ *tripping.Error) (shouldTrip bool) {
					return false
				},
				OpenDuration:                       0,
				OnStateChange:                      stateChange,
				HalfOpenSampler:                    samplerAlwaysSamples,
				NumberOfSuccessesInHalfOpenToClose: 10,
			})
			breaker.state = StateOpen
			breaker.openExpiresAt = time.Now()
			breaker.lastError = trippingError.Err
		})
		When("not expired", func() {
			BeforeEach(func() {
				breaker.openExpiresAt = time.Now().Add(1 * time.Hour)
			})
			It("does not transition", func() {
				_ = breaker.Use(func() error {
					return nil
				})
				Eventually(stateChange).ShouldNot(Receive())
			})
			It("returns the last error", func() {
				err := breaker.Use(func() error {
					return nil
				})
				Eventually(err).Should(Equal(trippingError.Err))
			})
		})
		When("expired", func() {
			It("transitions to half-open", func() {
				_ = breaker.Use(func() error {
					return nil
				})
				Expect(stateChange).Should(Receive(Equal(StateHalfOpen)))
			})
		})
	})
	When("half-open", func() {
		var (
			breaker *Breaker
		)
		BeforeEach(func() {
			breaker = New(Opts{
				TripDecider: func(_ *tripping.Error) (shouldTrip bool) {
					return false
				},
				OpenDuration:                       0,
				OnStateChange:                      stateChange,
				HalfOpenSampler:                    samplerAlwaysSamples,
				NumberOfSuccessesInHalfOpenToClose: 1,
			})
			breaker.state = StateHalfOpen
			breaker.halfOpenAt = time.Now()
			breaker.lastError = trippingError.Err
		})
		When("not sampled", func() {
			BeforeEach(func() {
				breaker.opts.HalfOpenSampler = samplerNeverSamples
			})
			It("does not transition", func() {
				_ = breaker.Use(func() error {
					return nil
				})
				Eventually(stateChange).ShouldNot(Receive())
			})
			It("returns the last error", func() {
				err := breaker.Use(func() error {
					return nil
				})
				Eventually(err).Should(Equal(trippingError.Err))
			})
		})
		When("sampled", func() {
			It("transitions to closed", func() {
				_ = breaker.Use(func() error {
					return nil
				})
				Expect(stateChange).Should(Receive(Equal(StateClosed)))
			})
		})
	})
})
