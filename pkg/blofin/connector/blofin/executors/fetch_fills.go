package executors

import (
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
)

const fillsHistoryPath = "/api/v1/trade/fills-history"

const FillsPageLimit = 100

func FetchAllFills(client *resty.Client, baseURL string, startMs int64) ([]models.Fill, error) {
	var result []models.Fill

	after := ""
	for {
		params := map[string]string{
			"limit": fmt.Sprintf("%d", FillsPageLimit),
		}
		if startMs > 0 {
			params["begin"] = fmt.Sprint(startMs)
		}
		if after != "" {
			params["after"] = after
		}

		page, err := doWithRateLimit(func() ([]models.Fill, error) {
			return blofin.DoGet[[]models.Fill](client, baseURL, fillsHistoryPath, params)
		})
		if err != nil {
			if after != "" && isHTTP5xx(err) {
				break
			}
			return nil, err
		}
		if len(page) == 0 {
			break
		}

		reachedCutoff := false
		for _, fill := range page {
			if startMs > 0 {
				ts, _ := strconv.ParseInt(fill.Ts, 10, 64)
				if ts > 0 && ts < startMs {
					reachedCutoff = true
					continue
				}
			}
			result = append(result, fill)
		}
		if reachedCutoff {
			break
		}

		if len(page) < FillsPageLimit {
			break
		}
		after = page[len(page)-1].TradeID
	}

	return result, nil
}

func FetchFillsPage(client *resty.Client, baseURL, after string, limit int) ([]models.Fill, error) {
	if limit <= 0 {
		limit = FillsPageLimit
	}

	params := map[string]string{
		"limit": fmt.Sprintf("%d", limit),
	}
	if after != "" {
		params["after"] = after
	}

	return doWithRateLimit(func() ([]models.Fill, error) {
		return blofin.DoGet[[]models.Fill](client, baseURL, fillsHistoryPath, params)
	})
}
