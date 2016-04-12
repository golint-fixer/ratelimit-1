// Package ratelimit provides an efficient token bucket implementation
// that can be used to limit the rate concurrent HTTP traffic.
// See http://en.wikipedia.org/wiki/Token_bucket.
package ratelimit

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/juju/ratelimit"
	"gopkg.in/vinxi/layer.v0"
)

// Filter represents the Limiter filter function signature.
type Filter func(r *http.Request) bool

// RateLimitResponder is used as default function to repond when the
// rate limit is reached. You can customize it via Limiter.SetResponder(fn).
var RateLimitResponder = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(429)
	w.Write([]byte("Too Many Requests"))
}

// Limiter implements a token bucket rate limiter middleware.
// Rate limiter can support multiple rate limit strategies, such as time based limiter.
type Limiter struct {
	// timeWindow stores the rate limiter time window.
	timeWindow time.Duration
	// lm provides thread synchronization to lastAvailable field.
	lm sync.Mutex
	// lastAvailable stores the last time that the limiter had available tokens.
	lastAvailable time.Time
	// bucket stores the ratelimit.Bucket limiter currently used.
	bucket *ratelimit.Bucket
	// responser stores the responder function used when the rate limit is reached.
	responder http.HandlerFunc
	// filters stores a list of filters to determine if should not apply the rate limiter.
	filters []Filter
}

// NewTimeLimiter creates a new time based rate limiter middleware.
func NewTimeLimiter(timeWindow time.Duration, capacity int64) *Limiter {
	return &Limiter{
		responder:     RateLimitResponder,
		timeWindow:    timeWindow,
		lastAvailable: time.Now(),
		bucket:        ratelimit.NewBucket(timeWindow, capacity),
	}
}

// NewRateLimiter creates a rate limiter middleware which limits the
// amount of requests accepted per second.
func NewRateLimiter(rate float64, capacity int64) *Limiter {
	return &Limiter{
		responder:  RateLimitResponder,
		timeWindow: time.Second,
		bucket:     ratelimit.NewBucketWithRate(rate, capacity),
	}
}

// SetResponder sets a custom function to reply in case of rate limit reached.
func (l *Limiter) SetResponder(fn http.HandlerFunc) {
	l.responder = fn
}

// AddFilter adds a new rate limiter whitelist filter.
// If the filter matches, the traffic won't be limited.
func (l *Limiter) AddFilter(fn ...Filter) {
	l.filters = append(l.filters, fn...)
}

// Register registers the middleware handler.
func (l *Limiter) Register(mw layer.Middleware) {
	mw.UsePriority("request", layer.TopHead, l.LimitHTTP)
}

// LimitHTTP limits an incoming HTTP request.
// If some filter passes, the request won't be limited.
// This method is used internally, but made public for public testing.
func (l *Limiter) LimitHTTP(h http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pass filters
		for _, filter := range l.filters {
			if filter(r) {
				h.ServeHTTP(w, r)
				return
			}
		}

		// Otherwise apply the rate limiter
		available := l.bucket.TakeAvailable(1)

		headers := w.Header()
		headers.Set("X-RateLimit-Limit", strconv.Itoa(l.capacity()))
		headers.Set("X-RateLimit-Remaining", strconv.Itoa(l.reamining()))

		// If no more tokens available, reply with 429
		if available == 0 {
			headers.Set("X-RateLimit-Reset", strconv.Itoa(int(l.resetTime())))
			l.responder(w, r)
			return
		}

		// Otherwise track time and forward the request
		l.trackTime()
		h.ServeHTTP(w, r)
	}
}

// resetTime is used to calculate the pending reset time to wait.
func (l *Limiter) resetTime() time.Duration {
	l.lm.Lock()
	defer l.lm.Unlock()
	return time.Now().Sub(l.lastAvailable) / time.Second
}

// trackTime is used to track the last time that tokens were available.
func (l *Limiter) trackTime() {
	l.lm.Lock()
	l.lastAvailable = time.Now()
	l.lm.Unlock()
}

// capacity is used to read the current bucket capacity.
func (l *Limiter) capacity() int {
	return int(l.bucket.Capacity())
}

// remaining is used to read the current pending remaining available buckets.
func (l *Limiter) reamining() int {
	if remaining := int(l.bucket.Available()); remaining > 0 {
		return remaining
	}
	return 0
}
