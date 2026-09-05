package bybit

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	ipWindow     = 5 * time.Second
	ipSoftBudget = 400

	defaultEndpointLimit = 5

	headerLimit       = "X-Bapi-Limit"
	headerLimitStatus = "X-Bapi-Limit-Status"
	headerLimitReset  = "X-Bapi-Limit-Reset-Timestamp"
)

type endpointBucket struct {
	limit     int
	remaining int
	windowEnd time.Time
	resetAt   time.Time
}

type RateLimiter struct {
	mu        sync.Mutex
	endpoints map[string]*endpointBucket
	counts    map[string]int
	ipStamps  []time.Time
	requests  int
	throttled int
}

var Limiter = NewRateLimiter()

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		endpoints: make(map[string]*endpointBucket),
		counts:    make(map[string]int),
	}
}

func (l *RateLimiter) Acquire(path string) {
	for {
		l.mu.Lock()
		now := time.Now()
		wait := l.ipWait(now)
		if wait == 0 {
			wait = l.endpointWait(path, now)
		}
		if wait == 0 {
			l.ipStamps = append(l.ipStamps, now)
			l.requests++
			l.mu.Unlock()
			return
		}
		l.mu.Unlock()
		time.Sleep(wait)
	}
}

func (l *RateLimiter) ipWait(now time.Time) time.Duration {
	cutoff := now.Add(-ipWindow)
	i := 0
	for i < len(l.ipStamps) && l.ipStamps[i].Before(cutoff) {
		i++
	}
	l.ipStamps = l.ipStamps[i:]
	if len(l.ipStamps) < ipSoftBudget {
		return 0
	}
	return l.ipStamps[0].Add(ipWindow).Sub(now) + 10*time.Millisecond
}

func (l *RateLimiter) endpointWait(path string, now time.Time) time.Duration {
	if isPublicPath(path) {
		l.counts[path]++
		return 0
	}
	l.counts[path]++
	b, ok := l.endpoints[path]
	if !ok {
		b = &endpointBucket{limit: defaultEndpointLimit, remaining: defaultEndpointLimit}
		l.endpoints[path] = b
	}
	if now.After(b.windowEnd) {
		b.windowEnd = now.Add(time.Second)
		b.remaining = b.limit
	}
	if !b.resetAt.IsZero() && now.Before(b.resetAt) {
		return b.resetAt.Sub(now) + 20*time.Millisecond
	}
	if b.remaining <= 0 {
		return b.windowEnd.Sub(now) + 20*time.Millisecond
	}
	b.remaining--
	return 0
}

func (l *RateLimiter) Observe(path string, h http.Header) {
	limit, limitOK := atoi(h.Get(headerLimit))
	status, statusOK := atoi(h.Get(headerLimitStatus))
	if !limitOK && !statusOK {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.endpoints[path]
	if !ok {
		b = &endpointBucket{}
		l.endpoints[path] = b
	}
	if limitOK && limit > 0 {
		b.limit = limit
	}
	if statusOK {
		if status < b.remaining || b.windowEnd.IsZero() {
			b.remaining = status
		}
		if status == 0 {
			l.throttled++
			resetMs, _ := strconv.ParseInt(h.Get(headerLimitReset), 10, 64)
			resetAt := time.UnixMilli(resetMs)
			if resetMs == 0 || !resetAt.After(time.Now()) {
				resetAt = time.Now().Add(200 * time.Millisecond)
			}
			b.resetAt = resetAt
		} else {
			b.resetAt = time.Time{}
		}
	}
}

func (l *RateLimiter) ResetAt(path string) time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()
	if b, ok := l.endpoints[path]; ok && !b.resetAt.IsZero() {
		return b.resetAt
	}
	return time.Now().Add(time.Second)
}

func (l *RateLimiter) Requests() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.requests
}

func (l *RateLimiter) Throttled() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.throttled
}

func (l *RateLimiter) Limits() map[string]int {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make(map[string]int, len(l.endpoints))
	for path, b := range l.endpoints {
		out[path] = b.limit
	}
	return out
}

func (l *RateLimiter) Counts() map[string]int {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make(map[string]int, len(l.counts))
	for path, n := range l.counts {
		out[path] = n
	}
	return out
}

func atoi(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	v, err := strconv.Atoi(s)
	return v, err == nil
}
