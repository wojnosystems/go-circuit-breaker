package twoStateCircuit

import (
	"github.com/wojnosystems/go-circuit-breaker/circuitTripping"
	"sync"
	"time"
)

type State uint8

const (
	StateClosed State = iota
	StateOpen
)

type Opts struct {
	FailureThreshold uint
	OpenDuration     time.Duration

	OnStateChange chan<- State
}

// Breaker is a live circuit breaker that only has 2 states: closed and open
type Breaker struct {
	breaker

	opts          Opts
	mu            sync.Mutex
	state         State
	errorCount    uint
	openExpiresAt time.Time
	lastError     error
	nowFactory    nowFactory
}

func New(opts Opts) *Breaker {
	return &Breaker{
		opts: opts,
	}
}

func (b *Breaker) now() time.Time {
	if b.nowFactory != nil {
		return b.nowFactory()
	}
	return time.Now()
}

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
	b.errorCount = 0
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

	b.errorCount++
	if b.errorCount >= b.opts.FailureThreshold {
		b.state = StateOpen
		b.openExpiresAt = b.now().Add(b.opts.OpenDuration)
		b.lastError = err.(*circuitTripping.Error).Unwrap()
		b.notifyStateChanged()
	}
}
