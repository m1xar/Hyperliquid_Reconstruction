package executors

import (
	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
)

const (
	closedPnlPath      = "/v5/position/closed-pnl"
	ClosedPnlPageLimit = 100
)

func FetchClosedPnlWindow(client *resty.Client, w Window) ([]models.ClosedPnl, error) {
	params := windowParams(w, map[string]string{"category": models.CategoryLinear})
	return collectCursor[models.ClosedPnl](client, closedPnlPath, params, ClosedPnlPageLimit)
}

func FetchClosedPnl(client *resty.Client, startMs, endMs int64) ([]models.ClosedPnl, error) {
	return ForEachWindow(Windows(startMs, endMs), DefaultWindowWorkers, func(w Window) ([]models.ClosedPnl, error) {
		return FetchClosedPnlWindow(client, w)
	})
}
