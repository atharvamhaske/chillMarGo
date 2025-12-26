package ratelimiter

import (
	"errors"
	"sync"
	"time"
)

var TooManyReqError = errors.New("Too many requests..")

type Limiter struct {
	buckets    map[string]*TokenBucket
	mutex      sync.Mutex
	capacity   int
	refillRate int
	now        func() time.Time
}

type TokenBucket struct {
	capacity       int
	refillRate     int
	tokens         int
	lastRefillTime time.Time
	lastSeen       time.Time
	mutex          sync.Mutex
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// newlimiter creates a limiter where each key gets its own token bucket
func NewLimiter(capacity, refillRate int) *Limiter {
	return &Limiter{
		buckets:    make(map[string]*TokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
		now:        time.Now,
	}
}
func NewTokenBucket(capacity, refillRate int, now time.Time) *TokenBucket {
	return &TokenBucket{
		capacity: capacity,
		refillRate: refillRate,
		tokens: capacity,
		lastRefillTime: now,
		lastSeen: now,
	}
}

func (tb *TokenBucket) refill(now time.Time) {
	elapsed := now.Sub(tb.lastRefillTime)
	seconds := int(elapsed / time.Second)

	if seconds <= 0 {
		return
	}

	tb.tokens = minInt(tb.tokens+seconds*tb.refillRate, tb.capacity)
	tb.lastRefillTime = tb.lastRefillTime.Add(time.Duration(seconds) * time.Second)
}

func (l *Limiter) getOrCreateKey(key string) *TokenBucket {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	tb, ok := l.buckets[key]
	if !ok {
		tb = NewTokenBucket(l.capacity, l.refillRate, l.now())
		l.buckets[key] = tb
	}
	tb.lastSeen = l.now()
	return tb
}

func (tb *TokenBucket) Take(tokensRequested int) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.refill(time.Now())

	if tokensRequested <= tb.tokens {
		tb.tokens -= tokensRequested
		return true
	}
	return false
}

func (tb *TokenBucket) Snapshot() (remaining, limit, retryAfterSeconds int) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.refill(time.Now())

	limit = tb.capacity
	remaining = tb.tokens

	if remaining > 0 {
		return remaining, limit, 0
	}

	// conservative retry-after
	if tb.refillRate > 0 {
		retryAfterSeconds = 1
	}

	return remaining, limit, retryAfterSeconds
}

func (l *Limiter) Allow(key string) bool {
	tb := l.getOrCreateKey(key)
	return tb.Take(1)
}

func (l *Limiter) Snapshot(key string) (remaining, limit, retryAfterSeconds int) {
	tb := l.getOrCreateKey(key)
	return tb.Snapshot()
}

func TokenBucketAlgo(l *Limiter) func(string) error {
	return func(key string) error {
		if l.Allow(key) {
			return nil
		}
		return TooManyReqError
	}
}

// TokenBucketAlgo adapts Limiter to func(string) error middleware style.