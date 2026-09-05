package binance

import (
	"strconv"
	"sync"
	"time"
)

const (
	weightLimitPerMinute = 2400
	weightSoftBudget     = 2000
)

type WeightLimiter struct {
	mu     sync.Mutex
	minute int64
	used   int
}

var Limiter = &WeightLimiter{}

func (l *WeightLimiter) Acquire(weight int) {
	if weight <= 0 {
		weight = 1
	}
	for {
		l.mu.Lock()
		now := time.Now()
		minute := now.Unix() / 60
		if minute != l.minute {
			l.minute = minute
			l.used = 0
		}
		if l.used+weight <= weightSoftBudget {
			l.used += weight
			l.mu.Unlock()
			return
		}
		l.mu.Unlock()

		nextMinute := time.Unix((minute+1)*60, 0)
		time.Sleep(time.Until(nextMinute) + 100*time.Millisecond)
	}
}

func (l *WeightLimiter) Observe(header string) {
	if header == "" {
		return
	}
	used, err := strconv.Atoi(header)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	minute := time.Now().Unix() / 60
	if minute != l.minute {
		l.minute = minute
		l.used = 0
	}
	if used > l.used {
		l.used = used
	}
}

func (l *WeightLimiter) Used() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.used
}
