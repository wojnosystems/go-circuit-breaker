package twoStateCircuit

import (
	"github.com/wojnosystems/go-circuit-breaker/circuitTripping"
	"github.com/wojnosystems/go-circuit-breaker/timeFactory"
	"sync"
	"time"
)

type Opts struct {
	// FailureThreshold how many times to fail in a row before opening the circuit
	FailureThreshold uint

	// FailureWindow length of time to record errors to consider whether the FailureThreshold has been met
	// errors older than this are discarded
	FailureWindow time.Duration

	// OpenDuration is how long to stay in the open state before closing again
	OpenDuration time.Duration

	// OnStateChange if set, will emit the state the breaker is transitioning into
	OnStateChange chan<- State
}

// Breaker is a live circuit breaker that only has 2 states: closed and open
type Breaker struct {
	breaker

	opts  Opts
	mu    sync.Mutex
	state State

	errorsOccurredAt []time.Time
	openExpiresAt    time.Time
	lastError        error
	nowFactory       timeFactory.Now
}

func New(opts Opts) *Breaker {
	return &Breaker{
		opts:             opts,
		errorsOccurredAt: make([]time.Time, 0, opts.FailureThreshold),
	}
}

func (b *Breaker) now() time.Time {
	if b.nowFactory != nil {
		return b.nowFactory()
	}
	return time.Now()
}

// Use the breaker, if closed, attempt the callback, if open, return the last error
// automatically transitions state if necessary
func (b *Breaker) Use(callback func() error) error {
	if err := b.reCloseOrFailFast(); err != nil {
		return err
	}

	err := callback()
	b.countErrorAndOpenIfNeeded(err)

	return err
}

func (b *Breaker) reCloseOrFailFast() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == StateClosed {
		return nil
	}
	if b.openExpiresAt.After(b.now()) {
		// Still open, return the last error
		return b.lastError
	}
	b.state = StateClosed
	b.errorsOccurredAt = b.errorsOccurredAt[0:0]
	b.notifyStateChanged()
	return nil
}

func (b *Breaker) notifyStateChanged() {
	if b.opts.OnStateChange != nil {
		b.opts.OnStateChange <- b.state
	}
}

func (b *Breaker) countErrorAndOpenIfNeeded(err error) {
	if !circuitTripping.IsTripping(err) {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state != StateClosed {
		// Not closed
		return
	}

	now := b.now()
	b.errorsOccurredAt = removeExpiredErrors(b.errorsOccurredAt, b.opts.FailureWindow, now)
	b.errorsOccurredAt = append(b.errorsOccurredAt, now)
	if uint(len(b.errorsOccurredAt)) >= b.opts.FailureThreshold {
		b.state = StateOpen
		b.openExpiresAt = b.now().Add(b.opts.OpenDuration)
		b.lastError = err.(*circuitTripping.Error).Unwrap()
		b.notifyStateChanged()
	}
}

func removeExpiredErrors(errorsOccurredIn []time.Time, failureWindow time.Duration, now time.Time) []time.Time {
	invalidBefore := now.Add(-1 * failureWindow)
	lastInvalidIndex := -1
	for i, t := range errorsOccurredIn {
		if t.Before(invalidBefore) {
			lastInvalidIndex = i
		}
		break
	}
	if lastInvalidIndex == -1 {
		return errorsOccurredIn
	}
	copy(errorsOccurredIn[:], errorsOccurredIn[lastInvalidIndex:])
	return errorsOccurredIn
}
