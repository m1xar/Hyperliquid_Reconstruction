package executors

import (
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
)

const ordersHistoryPath = "/api/v1/trade/orders-history"

const ordersPageLimit = 100

func FetchAllOrders(client *resty.Client, baseURL string, startMs int64) ([]models.Order, error) {
	var result []models.Order

	after := ""
	for {
		params := map[string]string{
			"limit": fmt.Sprintf("%d", ordersPageLimit),
		}
		if startMs > 0 {
			params["begin"] = fmt.Sprint(startMs)
		}
		if after != "" {
			params["after"] = after
		}

		page, err := doWithRateLimit(func() ([]models.Order, error) {
			return blofin.DoGet[[]models.Order](client, baseURL, ordersHistoryPath, params)
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
		for _, ord := range page {
			if startMs > 0 {
				updated, _ := strconv.ParseInt(ord.UpdateTime, 10, 64)
				if updated > 0 && updated < startMs {
					reachedCutoff = true
					continue
				}
			}
			result = append(result, ord)
		}
		if reachedCutoff {
			break
		}

		if len(page) < ordersPageLimit {
			break
		}
		after = page[len(page)-1].OrderID
	}

	return result, nil
}
