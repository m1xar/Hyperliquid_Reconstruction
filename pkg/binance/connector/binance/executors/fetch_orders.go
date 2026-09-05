package executors

import (
	"fmt"
	"sort"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/models"
)

const allOrdersPath = "/fapi/v1/allOrders"

const ordersPageLimit = 1000

func FetchOrdersFrom(client *resty.Client, symbol string, fromOrderID int64) ([]models.Order, error) {
	orders, err := fetchOrdersByID(client, symbol, fromOrderID)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(orders, func(i, j int) bool {
		if orders[i].Time == orders[j].Time {
			return orders[i].OrderID < orders[j].OrderID
		}
		return orders[i].Time < orders[j].Time
	})
	return orders, nil
}

func fetchOrdersByID(client *resty.Client, symbol string, fromOrderID int64) ([]models.Order, error) {
	var result []models.Order
	seen := make(map[int64]struct{})

	for {
		params := map[string]string{
			"symbol":  symbol,
			"orderId": fmt.Sprint(fromOrderID),
			"limit":   fmt.Sprintf("%d", ordersPageLimit),
		}

		page, err := doWithRateLimit(func() ([]models.Order, error) {
			return binance.DoGet[[]models.Order](client, allOrdersPath, params, 5)
		})
		if err != nil {
			if len(result) > 0 && isHTTP5xx(err) {
				break
			}
			return nil, err
		}
		if len(page) == 0 {
			break
		}

		added := 0
		maxID := int64(0)
		for _, ord := range page {
			if ord.OrderID > maxID {
				maxID = ord.OrderID
			}
			if _, ok := seen[ord.OrderID]; ok {
				continue
			}
			seen[ord.OrderID] = struct{}{}
			result = append(result, ord)
			added++
		}

		if len(page) < ordersPageLimit || added == 0 || maxID < fromOrderID {
			break
		}
		fromOrderID = maxID + 1
	}

	return result, nil
}
