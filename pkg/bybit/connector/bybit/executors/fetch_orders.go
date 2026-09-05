package executors

import (
	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
)

const (
	orderHistoryPath = "/v5/order/history"
	OrdersPageLimit  = 50
)

func FetchOrdersWindow(client *resty.Client, w Window) ([]models.Order, error) {
	params := windowParams(w, map[string]string{"category": models.CategoryLinear})
	return collectCursor[models.Order](client, orderHistoryPath, params, OrdersPageLimit)
}

func FetchOrders(client *resty.Client, startMs, endMs int64) ([]models.Order, error) {
	rows, err := ForEachWindow(Windows(startMs, endMs), DefaultWindowWorkers, func(w Window) ([]models.Order, error) {
		return FetchOrdersWindow(client, w)
	})
	if err != nil {
		return nil, err
	}
	return DedupeOrders(rows), nil
}

func DedupeOrders(rows []models.Order) []models.Order {
	seen := make(map[string]struct{}, len(rows))
	out := rows[:0]
	for _, row := range rows {
		if _, ok := seen[row.OrderID]; ok {
			continue
		}
		seen[row.OrderID] = struct{}{}
		out = append(out, row)
	}
	return out
}
