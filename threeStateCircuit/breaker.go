package threeStateCircuit

import (
	"github.com/wojnosystems/go-circuit-breaker/tripping"
	"github.com/wojnosystems/go-time-factory/timeFactory"
	"sync"
	"time"
)

type Opts struct {
	// TripDecider is consulted each time a tripping error occurs.
	TripDecider tripping.Decider

	// OpenDuration is how long to stay in the OpenState before closing again
	OpenDuration time.Duration

	// OnStateChange if set, will emit the state the breaker is transitioning into
	// leaving as nil to avoid listening to state changes
	// Do NOT close this channel or a panic will occur
	OnStateChange chan<- State

	// HalfOpenSampler tells the circuit breaker which requests to reject and which to attempt while in the half-open state
	HalfOpenSampler ShouldSample

	// NumberOfSuccessesInHalfOpenToClose is the number of times the requests need to succeed while in the Half-Open state
	// in order to transition back to the closed state. Any error in the half-open state, will reset it back to the open state
	NumberOfSuccessesInHalfOpenToClose uint64

	// nowFactory allows the current time to be simulated
	nowFactory timeFactory.Now
}

type mutableState struct {
	state             State
	lastError         error
	openExpiresAt     time.Time
	halfOpenAt        time.Time
	halfOpenSuccesses uint64
}

// Breaker is a live circuit breaker that only has 2 states: closed and open.
// Use New to create a new Breaker, populated with options.
type Breaker struct {
	opts Opts
	mu   sync.RWMutex
	mutableState
}

func New(opts Opts) *Breaker {
	return &Breaker{
		opts: opts,
		mutableState: mutableState{
			state: StateClosed,
		},
	}
}

// Use the breaker, if closed, attempt the callback, if open, return the last error
// automatically transitions state if necessary
func (b *Breaker) Use(callback func() error) error {
	stateCopy, now := b.copyCurrentState()
	if stateCopy.state == StateOpen {
		if stateCopy.openExpiresAt.After(now) {
			// still in the open state, not expired
			return stateCopy.lastError
		}

		stateCopy = b.transitionToHalfOpenIfShould()
	}

	if stateCopy.state == StateHalfOpen {
		if !b.opts.HalfOpenSampler.ShouldSample(b.opts.nowFactory.Get().Sub(stateCopy.halfOpenAt)) {
			return b.lastError
		}
	}

	// at this point, we have either returned or we're in the closed state
	err := callback()
	if !tripping.IsTripping(err) {
		b.mu.RLock()
		currentState := b.state
		b.mu.RUnlock()
		if currentState == StateHalfOpen {
			b.recordSuccessAndTransitionToClosedIfShould()
		}
		// error was nil or not tripping, just return
		return err
	}

	trippingError := err.(*tripping.Error)
	unwrappedError := err.(*tripping.Error).Err

	// we encountered an error, we need to count this against our error threshold and transition if need be
	b.recordErrorAndTransitionToOpenIfShould(trippingError)
	return unwrappedError
}

func (b *Breaker) copyCurrentState() (currentState mutableState, now time.Time) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	currentState.state = b.state
	currentState.openExpiresAt = b.openExpiresAt
	currentState.lastError = b.lastError
	now = b.opts.nowFactory.Get()
	return
}

func doNothing() {}

func (b *Breaker) transitionToHalfOpenIfShould() mutableState {
	afterUnlock := doNothing
	b.mu.Lock()
	defer func() {
		b.mu.Unlock()
		afterUnlock()
	}()
	// are we still recorded as being in the open state?
	if b.state == StateOpen {
		// perform the transition exactly once for this round
		b.state = StateHalfOpen
		b.halfOpenAt = b.opts.nowFactory.Get()
		b.halfOpenSuccesses = 0
		afterUnlock = func() {
			b.notifyStateChanged(StateHalfOpen)
		}
	}
	return b.mutableState
}

func (b *Breaker) recordErrorAndTransitionToOpenIfShould(trippingError *tripping.Error) {
	b.mu.Lock()
	afterUnlock := doNothing
	defer func() {
		b.mu.Unlock()
		afterUnlock()
	}()

	if b.state == StateClosed {
		// record the error
		errorRateWithinLimits := !b.opts.TripDecider.ShouldTrip(trippingError)
		if errorRateWithinLimits {
			return
		}
	}

	if b.state == StateOpen {
		// already transitioned state to open OR
		// error rate not yet exceeded, no need to transition
		return
	}

	// transition to the Open State
	b.lastError = trippingError.Err
	b.state = StateOpen
	b.openExpiresAt = b.opts.nowFactory.Get().Add(b.opts.OpenDuration)
	afterUnlock = func() {
		b.notifyStateChanged(StateOpen)
	}
}

func (b *Breaker) notifyStateChanged(newState State) {
	if b.opts.OnStateChange != nil {
		b.opts.OnStateChange <- newState
	}
}

func (b *Breaker) recordSuccessAndTransitionToClosedIfShould() {
	afterUnlock := doNothing
	b.mu.Lock()
	defer func() {
		b.mu.Unlock()
		afterUnlock()
	}()
	// are we still recorded as being in the open state?
	if b.state == StateHalfOpen {
		b.halfOpenSuccesses++
		if b.halfOpenSuccesses >= b.opts.NumberOfSuccessesInHalfOpenToClose {
			// perform the transition exactly once for this round
			b.state = StateClosed
			afterUnlock = func() {
				b.notifyStateChanged(StateClosed)
			}
		}
	}
}
