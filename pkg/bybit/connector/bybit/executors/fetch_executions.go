package executors

import (
	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
)

const (
	executionListPath   = "/v5/execution/list"
	ExecutionsPageLimit = 100
)

func FetchExecutionsWindow(client *resty.Client, w Window, symbol string) ([]models.Execution, error) {
	params := windowParams(w, map[string]string{
		"category": models.CategoryLinear,
		"symbol":   symbol,
	})
	return collectCursor[models.Execution](client, executionListPath, params, ExecutionsPageLimit)
}
