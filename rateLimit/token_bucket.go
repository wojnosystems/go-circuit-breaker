package rateLimit

import (
	"github.com/wojnosystems/go-circuit-breaker/timeFactory"
	"time"
)

var (
	zeroTime = time.Time{}
)

func defaultNow() time.Time {
	return time.Now()
}

type TokenBucketOpts struct {
	MaxTokens            uint64
	TokensAddedPerSecond float64
	InitialTokens        uint64
}

type TokenBucket struct {
	opts TokenBucketOpts

	tokens      uint64
	lastUpdated time.Time

	nowFactory timeFactory.Now
}

func NewTokenBucket(opts TokenBucketOpts) *TokenBucket {
	return &TokenBucket{
		opts: opts,
	}
}

// Allowed records the use and returns true if the rate limit is not exceeded.
func (b *TokenBucket) Allowed() bool {
	b.initializeIfZero()
	b.replenish()
	if b.tokens != 0 {
		b.tokens--
		return true
	}
	return false
}

func (b *TokenBucket) initializeIfZero() {
	if b.lastUpdated != zeroTime {
		return
	}

	if b.nowFactory == nil {
		b.nowFactory = defaultNow
	}

	b.tokens = b.opts.InitialTokens
	b.lastUpdated = b.nowFactory()
}

func (b *TokenBucket) replenish(now time.Time) {
	tokensToAdd := b.opts.TokensAddedPerSecond * now.Sub(b.lastUpdated).Seconds()
	b.tokens
}
