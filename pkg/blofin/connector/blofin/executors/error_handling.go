package executors

import (
	"errors"
	"time"

	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin"
)

const (
	rateLimitSleep   = 2 * time.Second
	rateLimitRetries = 3
)

func isHTTP5xx(err error) bool {
	var httpErr *blofin.HTTPError
	return errors.As(err, &httpErr) && httpErr.StatusCode >= 500
}

func isThrottled(err error) bool {
	var httpErr *blofin.HTTPError
	if !errors.As(err, &httpErr) {
		return false
	}
	return httpErr.StatusCode == 429 || httpErr.StatusCode == 403
}

func doWithRateLimit[T any](fn func() (T, error)) (T, error) {
	for i := 0; ; i++ {
		result, err := fn()
		if err != nil && isThrottled(err) && i < rateLimitRetries {
			time.Sleep(rateLimitSleep * time.Duration(i+1))
			continue
		}
		return result, err
	}
}
