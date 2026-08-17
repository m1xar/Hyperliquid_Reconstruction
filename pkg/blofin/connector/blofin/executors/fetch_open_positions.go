package executors

import (
	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
)

const positionsPath = "/api/v1/account/positions"

func FetchOpenPositions(client *resty.Client, baseURL string) ([]models.OpenPosition, error) {
	return doWithRateLimit(func() ([]models.OpenPosition, error) {
		return blofin.DoGet[[]models.OpenPosition](client, baseURL, positionsPath, nil)
	})
}
