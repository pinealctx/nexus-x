package agentic

import (
	"context"
	"sync"
	"time"
)

// InMemoryRateLimiter is a simple token-bucket rate limiter for development and testing.
type InMemoryRateLimiter struct {
	mu       sync.Mutex
	buckets  map[int32]*bucket
	rate     int           // max requests per window
	window   time.Duration // sliding window duration
}

type bucket struct {
	count    int
	resetAt  time.Time
}

// NewInMemoryRateLimiter creates a rate limiter that allows `rate` requests per `window`.
func NewInMemoryRateLimiter(rate int, window time.Duration) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		buckets: make(map[int32]*bucket),
		rate:    rate,
		window:  window,
	}
}

// Allow returns true if the user is within rate limits.
func (r *InMemoryRateLimiter) Allow(_ context.Context, userID int32) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b, ok := r.buckets[userID]
	if !ok || now.After(b.resetAt) {
		r.buckets[userID] = &bucket{count: 1, resetAt: now.Add(r.window)}
		return true, nil
	}
	if b.count >= r.rate {
		return false, nil
	}
	b.count++
	return true, nil
}

// InMemoryDedup is a simple in-memory deduplicator using a TTL-based cache.
type InMemoryDedup struct {
	mu      sync.Mutex
	seen    map[dedupKey]time.Time
	ttl     time.Duration
	maxSize int
}

type dedupKey struct {
	conversationID int64
	messageID      int64
}

// NewInMemoryDedup creates a deduplicator that remembers messages for the given TTL.
func NewInMemoryDedup(ttl time.Duration) *InMemoryDedup {
	return &InMemoryDedup{
		seen:    make(map[dedupKey]time.Time),
		ttl:     ttl,
		maxSize: 10000,
	}
}

// IsDuplicate returns true if this message has been seen within the TTL window.
func (d *InMemoryDedup) IsDuplicate(conversationID int64, messageID int64) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := dedupKey{conversationID, messageID}
	now := time.Now()

	if t, ok := d.seen[key]; ok && now.Before(t.Add(d.ttl)) {
		return true
	}

	// Evict expired entries if over capacity.
	if len(d.seen) >= d.maxSize {
		for k, t := range d.seen {
			if now.After(t.Add(d.ttl)) {
				delete(d.seen, k)
			}
		}
	}

	d.seen[key] = now
	return false
}
