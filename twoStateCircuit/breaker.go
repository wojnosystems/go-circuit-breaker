package twoStateCircuit

import (
	"github.com/wojnosystems/go-circuit-breaker/circuitTripping"
	"github.com/wojnosystems/go-time-factory/timeFactory"
	"sync"
	"time"
)

const errorCost = 1

type Limiter interface {
	Allowed(tokenCost uint64) bool
}

type Opts struct {
	// FailureLimiter is a rate limiter called each time an error occurs while in the ClosedState
	// After the rate is exceeded, the breaker will enter the OpenState.
	// This is required to prevent the breaker from remaining in the OpenState and allowing an error budget before
	// tripping. This must be set.
	// This limiter should not be shared with any other go routines as this needs to be locked to prevent race conditions
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

type postClose func(b *Breaker, err error) (afterUnlock func())
type postOpen func(b *Breaker) (afterUnlock func())

// Breaker is a live circuit breaker that only has 2 states: closed and open.
// Use New to create a new Breaker, populated with options.
type Breaker struct {
	opts                Opts
	mu                  sync.RWMutex
	state               stateFunc
	postCloseTripAction postClose
	postOpenResetAction postOpen

	openExpiresAt time.Time
	lastError     error
}

func New(opts Opts) *Breaker {
	return &Breaker{
		opts:                opts,
		state:               stateClosed,
		postCloseTripAction: postCloseTripTransitionAction,
		postOpenResetAction: postOpenResetTransitionAction,
	}
}

type stateContext struct {
	breaker  *Breaker
	callback func() error
}

type stateFunc func(b *stateContext) error

// Use the breaker, if closed, attempt the callback, if open, return the last error
// automatically transitions state if necessary
func (b *Breaker) Use(callback func() error) error {
	b.mu.RLock()
	stateF := b.state
	b.mu.RUnlock()
	return stateF(&stateContext{
		breaker:  b,
		callback: callback,
	})
}

func stateClosed(b *stateContext) error {
	err := b.callback()
	if !circuitTripping.IsTripping(err) {
		return err
	}
	afterUnlock := doNothing
	b.breaker.mu.Lock()
	defer func() {
		b.breaker.mu.Unlock()
		afterUnlock()
	}()
	afterUnlock = b.breaker.postCloseTripAction(b.breaker, err)
	return err
}

func doNothing() {}

func postCloseTripTransitionAction(b *Breaker, err error) (afterUnlock func()) {
	if !b.opts.FailureLimiter.Allowed(errorCost) {
		b.lastError = err.(*circuitTripping.Error).Unwrap()
		b.state = stateOpen
		b.postCloseTripAction = postCloseTripDoNothingAction
		b.postOpenResetAction = postOpenResetTransitionAction
		b.openExpiresAt = b.opts.nowFactory.Get().Add(b.opts.OpenDuration)
		return func() {
			b.notifyStateChanged(StateOpen)
		}
	}
	return doNothing
}

func postCloseTripDoNothingAction(_ *Breaker, _ error) (afterUnlock func()) {
	return doNothing
}

func (b *Breaker) notifyStateChanged(newState State) {
	if b.opts.OnStateChange != nil {
		b.opts.OnStateChange <- newState
	}
}

func stateOpen(b *stateContext) error {
	b.breaker.mu.RLock()
	if b.breaker.openExpiresAt.After(b.breaker.opts.nowFactory.Get()) {
		// Still open, return the last error
		b.breaker.mu.RUnlock()
		return b.breaker.lastError
	}
	b.breaker.mu.RUnlock()

	afterUnlock := doNothing
	b.breaker.mu.Lock()
	defer func() {
		b.breaker.mu.Unlock()
		afterUnlock()
	}()
	afterUnlock = b.breaker.postOpenResetAction(b.breaker)
	return nil
}

func postOpenResetTransitionAction(breaker *Breaker) (afterUnlock func()) {
	breaker.state = stateClosed
	breaker.postCloseTripAction = postCloseTripTransitionAction
	breaker.postOpenResetAction = postOpenDoNothingAction
	return func() {
		breaker.notifyStateChanged(StateClosed)
	}
}

func postOpenDoNothingAction(_ *Breaker) (afterUnlock func()) {
	return doNothing
}
