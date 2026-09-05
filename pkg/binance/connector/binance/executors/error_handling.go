package executors

import (
	"errors"
	"time"

	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance"
)

const (
	rateLimitSleep   = 2 * time.Second
	rateLimitRetries = 4
)

func asHTTPError(err error) (*binance.HTTPError, bool) {
	var httpErr *binance.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr, true
	}
	return nil, false
}

func isHTTP5xx(err error) bool {
	httpErr, ok := asHTTPError(err)
	return ok && httpErr.StatusCode >= 500
}

func isThrottled(err error) bool {
	httpErr, ok := asHTTPError(err)
	return ok && httpErr.StatusCode == 429
}

func isBanned(err error) bool {
	httpErr, ok := asHTTPError(err)
	return ok && httpErr.StatusCode == 418
}

func doWithRateLimit[T any](fn func() (T, error)) (T, error) {
	for i := 0; ; i++ {
		result, err := fn()
		if err == nil || isBanned(err) || i >= rateLimitRetries {
			return result, err
		}
		if isThrottled(err) {
			sleep := rateLimitSleep * time.Duration(i+1)
			if httpErr, ok := asHTTPError(err); ok && httpErr.RetryAfter > 0 {
				sleep = httpErr.RetryAfter + time.Second
			}
			time.Sleep(sleep)
			continue
		}
		if isHTTP5xx(err) {
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		return result, err
	}
}
