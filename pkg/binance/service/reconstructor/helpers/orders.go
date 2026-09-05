package helpers

import (
	"math"
	"strings"
	"time"

	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/models"
)

const tpslLookback = time.Minute

func IndexOrdersByID(orders []models.Order) map[int64]models.Order {
	idx := make(map[int64]models.Order, len(orders))
	for _, o := range orders {
		idx[o.OrderID] = o
	}
	return idx
}

func GroupOrdersBySymbol(orders []models.Order) map[string][]models.Order {
	idx := make(map[string][]models.Order)
	for _, o := range orders {
		idx[o.Symbol] = append(idx[o.Symbol], o)
	}
	return idx
}

func OrdersForEpisode(ep Episode, ordersByID map[int64]models.Order) map[int64]models.Order {
	out := make(map[int64]models.Order, len(ep.Parts))
	for _, part := range ep.Parts {
		if order, ok := ordersByID[part.Fill.OrderID]; ok {
			out[part.Fill.OrderID] = order
		}
	}
	return out
}

func TPSLForEpisode(ep Episode, symbolOrders []models.Order) (tp, sl *float64) {
	if len(symbolOrders) == 0 {
		return nil, nil
	}

	fromMs := ep.OpenAt.Add(-tpslLookback).UnixMilli()
	toMs := ep.CloseAt.UnixMilli()
	closingSide := "SELL"
	if ep.OpenSign < 0 {
		closingSide = "BUY"
	}

	var tpAt, slAt int64
	for _, ord := range symbolOrders {
		if ord.Symbol != ep.Symbol {
			continue
		}
		if !positionSideCompatible(ep.PositionSide, ord.PositionSide) {
			continue
		}
		if !strings.EqualFold(ord.Side, closingSide) {
			continue
		}
		if ord.Time < fromMs || ord.Time > toMs {
			continue
		}

		trigger := MustFloat(ord.StopPrice)
		if trigger <= 0 {
			continue
		}

		switch {
		case IsTakeProfitType(ord.OrigType):
			if tp == nil || ord.Time < tpAt {
				v := Round8(trigger)
				tp, tpAt = &v, ord.Time
			}
		case IsStopLossType(ord.OrigType):
			if sl == nil || ord.Time < slAt {
				v := Round8(trigger)
				sl, slAt = &v, ord.Time
			}
		}
	}
	return tp, sl
}

func positionSideCompatible(episodeSide, orderSide string) bool {
	e := strings.ToLower(strings.TrimSpace(episodeSide))
	o := strings.ToLower(strings.TrimSpace(orderSide))
	if e == "" || e == "both" || o == "" || o == "both" {
		return true
	}
	return e == o
}

func AllocateFillStats(part FillPart) (fee, pnl float64) {
	ratio := 1.0
	if fullSize := MustFloat(part.Fill.Qty); fullSize > sizeEpsilon {
		ratio = part.Size / fullSize
		if ratio > 1 {
			ratio = 1
		}
	}
	fee = math.Abs(MustFloat(part.Fill.Commission)) * ratio
	pnl = MustFloat(part.Fill.RealizedPnl) * ratio
	return fee, pnl
}
