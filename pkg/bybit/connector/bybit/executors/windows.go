package executors

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
)

const (
	WindowSpan = 7*24*time.Hour - time.Second

	HistoryRetention = 729 * 24 * time.Hour

	DefaultWindowWorkers = 4
)

type Window struct {
	StartMs int64
	EndMs   int64
}

func Windows(startMs, endMs int64) []Window {
	if endMs <= 0 {
		endMs = time.Now().UnixMilli()
	}
	floor := time.Now().Add(-HistoryRetention).UnixMilli()
	if startMs <= 0 || startMs < floor {
		startMs = floor
	}

	span := WindowSpan.Milliseconds()
	out := make([]Window, 0, int((endMs-startMs)/span)+1)
	for end := endMs; end >= startMs; {
		start := end - span
		if start < startMs {
			start = startMs
		}
		out = append(out, Window{StartMs: start, EndMs: end})
		end = start - 1
	}
	return out
}

func ForEachWindow[T any](windows []Window, workers int, fn func(w Window) ([]T, error)) ([]T, error) {
	if workers <= 0 {
		workers = DefaultWindowWorkers
	}

	results := make([][]T, len(windows))
	errs := make([]error, len(windows))
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for i, w := range windows {
		wg.Add(1)
		go func(i int, w Window) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i], errs[i] = fn(w)
		}(i, w)
	}
	wg.Wait()

	var out []T
	for i := range windows {
		if errs[i] != nil {
			return nil, errs[i]
		}
		out = append(out, results[i]...)
	}
	return out, nil
}

func collectCursor[T any](client *resty.Client, path string, params map[string]string, limit int) ([]T, error) {
	var result []T
	cursor := ""

	for {
		p := make(map[string]string, len(params)+2)
		for k, v := range params {
			p[k] = v
		}
		p["limit"] = fmt.Sprintf("%d", limit)
		if cursor != "" {
			p["cursor"] = cursor
		}

		page, err := doWithRateLimit(func() (models.CursorPage[T], error) {
			return bybit.DoGet[models.CursorPage[T]](client, path, p)
		})
		if err != nil {
			if isBeyondRetention(err) || (len(result) > 0 && isHTTP5xx(err)) {
				break
			}
			return nil, err
		}
		if len(page.List) == 0 {
			break
		}

		result = append(result, page.List...)

		if page.NextPageCursor == "" || page.NextPageCursor == cursor || len(page.List) < limit {
			break
		}
		cursor = page.NextPageCursor
	}

	return result, nil
}

func windowParams(w Window, extra map[string]string) map[string]string {
	p := map[string]string{
		"startTime": fmt.Sprint(w.StartMs),
		"endTime":   fmt.Sprint(w.EndMs),
	}
	for k, v := range extra {
		if v != "" {
			p[k] = v
		}
	}
	return p
}
