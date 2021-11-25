package twoStateCircuit_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/wojnosystems/go-circuit-breaker/circuitTripping"
	"github.com/wojnosystems/go-circuit-breaker/twoStateCircuit"
	"time"
)

var trippingError = circuitTripping.New(errors.New("tripping error"))

var _ = Describe("Breaker", func() {
	When("no errors", func() {
		var (
			subject     *twoStateCircuit.Breaker
			stateChange chan twoStateCircuit.State
		)
		BeforeEach(func() {
			stateChange = make(chan twoStateCircuit.State, 10)
			subject = twoStateCircuit.New(twoStateCircuit.Opts{
				FailureThreshold: 1,
				OpenDuration:     1 * time.Hour,
				OnStateChange:    stateChange,
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
		It("does not trip", func() {
			Expect(stateChange).ShouldNot(Receive())
		})
	})

	When("error threshold met", func() {
		var (
			subject     *twoStateCircuit.Breaker
			stateChange chan twoStateCircuit.State
			options     twoStateCircuit.Opts
		)
		BeforeEach(func() {
			stateChange = make(chan twoStateCircuit.State, 10)
			options = twoStateCircuit.Opts{
				FailureThreshold: 1,
				OpenDuration:     50 * time.Millisecond,
				OnStateChange:    stateChange,
			}
			subject = twoStateCircuit.New(options)

			_ = subject.Use(func() error {
				return trippingError
			})
		})
		It("returns the last error", func() {
			err := subject.Use(func() error {
				return nil
			})
			Expect(err).Should(HaveOccurred())
			Expect(err).Should(Equal(trippingError.Unwrap()))
		})
		It("opens", func() {
			Expect(stateChange).Should(Receive(Equal(twoStateCircuit.StateOpen)))
		})
		It("closes again", func() {
			<-time.After(options.OpenDuration + 1*time.Millisecond)
			_ = subject.Use(func() error {
				return nil
			})
			Expect(stateChange).Should(Receive(Equal(twoStateCircuit.StateOpen)))
			Expect(stateChange).Should(Receive(Equal(twoStateCircuit.StateClosed)))
		})
	})
})
