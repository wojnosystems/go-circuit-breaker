package twoStateCircuit

import (
	"github.com/wojnosystems/go-circuit-breaker/circuitTripping"
	"github.com/wojnosystems/go-time-factory/timeFactory"
	"sync"
	"time"
)

type Limiter interface {
	Allowed(tokenCost uint64) bool
}

type Opts struct {
	// FailureLimiter is a rate limiter called each time an error occurs while in the ClosedState
	// After the rate is exceeded, the breaker will enter the OpenState.
	// This is required to prevent the breaker from remaining in the OpenState and allowing an error budget before
	// tripping. This must be set.
	FailureLimiter Limiter

	// OpenDuration is how long to stay in the OpenState before closing again
	OpenDuration time.Duration

	// OnStateChange if set, will emit the state the breaker is transitioning into
	// leaving as nil to avoid listening to state changes
	// Do NOT close this channel or a panic will occur
	OnStateChange chan<- State

	// nowFactory allows the current time to be simulated
	nowFactory timeFactory.Now
}

// Breaker is a live circuit breaker that only has 2 states: closed and open
type Breaker struct {
	opts  Opts
	mu    sync.Mutex
	state State

	openExpiresAt time.Time
	lastError     error
}

func New(opts Opts) *Breaker {
	return &Breaker{
		opts: opts,
	}
}

// Use the breaker, if closed, attempt the callback, if open, return the last error
// automatically transitions state if necessary
func (b *Breaker) Use(callback func() error) (err error) {
	if err = b.reCloseOrFailFast(); err != nil {
		return
	}
	err = callback()
	b.countErrorAndOpenIfNeeded(err)
	return
}

func (b *Breaker) reCloseOrFailFast() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == StateClosed {
		return nil
	}
	if b.openExpiresAt.After(b.opts.nowFactory.Get()) {
		// Still open, return the last error
		return b.lastError
	}
	b.transitionClosed()
	return nil
}

func (b *Breaker) transitionClosed() {
	b.state = StateClosed
	b.notifyStateChanged(StateClosed)
}

func (b *Breaker) notifyStateChanged(newState State) {
	if b.opts.OnStateChange != nil {
		b.opts.OnStateChange <- newState
	}
}

func (b *Breaker) countErrorAndOpenIfNeeded(err error) {
	if !circuitTripping.IsTripping(err) {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.opts.FailureLimiter.Allowed(1) {
		b.lastError = err.(*circuitTripping.Error).Unwrap()
		b.transitionOpen()
	}
}

func (b *Breaker) transitionOpen() {
	b.state = StateOpen
	b.openExpiresAt = b.opts.nowFactory.Get().Add(b.opts.OpenDuration)
	b.notifyStateChanged(StateOpen)
}
