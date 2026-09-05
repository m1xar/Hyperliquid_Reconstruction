package executors

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit"
)

const (
	rateLimitRetries = 5
	serverErrorSleep = time.Second
)

func asHTTPError(err error) (*bybit.HTTPError, bool) {
	var httpErr *bybit.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr, true
	}
	return nil, false
}

func asAPIError(err error) (*bybit.APIError, bool) {
	var apiErr *bybit.APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

func isHTTP5xx(err error) bool {
	if httpErr, ok := asHTTPError(err); ok {
		return httpErr.StatusCode >= 500
	}
	if apiErr, ok := asAPIError(err); ok {
		return apiErr.Code == bybit.CodeServerError
	}
	return false
}

func isThrottled(err error) bool {
	if httpErr, ok := asHTTPError(err); ok {
		return httpErr.StatusCode == http.StatusTooManyRequests
	}
	if apiErr, ok := asAPIError(err); ok {
		return apiErr.Code == bybit.CodeRateLimit || apiErr.Code == bybit.CodeIPRateLimit
	}
	return false
}

func isBeyondRetention(err error) bool {
	apiErr, ok := asAPIError(err)
	return ok && apiErr.Code == 10001 && strings.Contains(strings.ToLower(apiErr.Msg), "2 years")
}

func isBanned(err error) bool {
	httpErr, ok := asHTTPError(err)
	return ok && httpErr.IPBanned()
}

func throttleSleep(err error, attempt int) time.Duration {
	if apiErr, ok := asAPIError(err); ok && !apiErr.ResetAt.IsZero() {
		if d := time.Until(apiErr.ResetAt); d > 0 {
			return d + 50*time.Millisecond
		}
	}
	if httpErr, ok := asHTTPError(err); ok && httpErr.RetryAfter > 0 {
		return httpErr.RetryAfter + time.Second
	}
	return time.Duration(attempt+1) * 500 * time.Millisecond
}

func doWithRateLimit[T any](fn func() (T, error)) (T, error) {
	for i := 0; ; i++ {
		result, err := fn()
		if err == nil || isBanned(err) || i >= rateLimitRetries {
			return result, err
		}
		if isThrottled(err) {
			time.Sleep(throttleSleep(err, i))
			continue
		}
		if isHTTP5xx(err) {
			time.Sleep(serverErrorSleep * time.Duration(i+1))
			continue
		}
		return result, err
	}
}
