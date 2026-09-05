package helpers

import (
	"math"
	"strings"
	"time"

	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
)

const tpslLookback = time.Minute

func IndexOrdersByID(orders []models.Order) map[string]models.Order {
	idx := make(map[string]models.Order, len(orders))
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

func IndexClosedPnlByOrder(rows []models.ClosedPnl) map[string]models.ClosedPnl {
	idx := make(map[string]models.ClosedPnl, len(rows))
	for _, r := range rows {
		idx[r.OrderID] = r
	}
	return idx
}

func OrdersForEpisode(ep Episode, ordersByID map[string]models.Order) map[string]models.Order {
	out := make(map[string]models.Order, len(ep.Parts))
	for _, part := range ep.Parts {
		if order, ok := ordersByID[part.Fill.OrderID()]; ok {
			out[part.Fill.OrderID()] = order
		}
	}
	return out
}

func LeverageForEpisode(ep Episode, closedByOrder map[string]models.ClosedPnl) float64 {
	for i := len(ep.Parts) - 1; i >= 0; i-- {
		part := ep.Parts[i]
		if part.Sign == ep.OpenSign {
			continue
		}
		if row, ok := closedByOrder[part.Fill.OrderID()]; ok {
			if lev := MustFloat(row.Leverage); lev > 0 {
				return lev
			}
		}
	}
	return 0
}

func TPSLForEpisode(ep Episode, episodeOrders map[string]models.Order, symbolOrders []models.Order) (tp, sl *float64) {
	var tpAt, slAt int64 = math.MaxInt64, math.MaxInt64

	consider := func(kind string, price float64, at int64) {
		if price <= 0 {
			return
		}
		v := Round8(price)
		switch kind {
		case "tp":
			if tp == nil || at < tpAt {
				tp, tpAt = &v, at
			}
		case "sl":
			if sl == nil || at < slAt {
				sl, slAt = &v, at
			}
		}
	}

	for _, part := range ep.Parts {
		if part.Sign != ep.OpenSign {
			continue
		}
		order, ok := episodeOrders[part.Fill.OrderID()]
		if !ok {
			continue
		}
		at := order.CreatedTime.Int64()
		consider("tp", MustFloat(order.TakeProfit), at)
		consider("sl", MustFloat(order.StopLoss), at)
	}
	if tp != nil && sl != nil {
		return tp, sl
	}

	fromMs := ep.OpenAt.Add(-tpslLookback).UnixMilli()
	toMs := ep.CloseAt.UnixMilli()
	closingSide := "Sell"
	if ep.OpenSign < 0 {
		closingSide = "Buy"
	}

	for _, ord := range symbolOrders {
		if ord.Symbol != ep.Symbol || !strings.EqualFold(ord.Side, closingSide) {
			continue
		}
		if !positionIdxCompatible(ep.PositionIdx, int(ord.PositionIdx.Int64())) {
			continue
		}
		at := ord.CreatedTime.Int64()
		if at < fromMs || at > toMs {
			continue
		}
		trigger := MustFloat(ord.TriggerPrice)
		switch {
		case IsTakeProfitType(ord.StopOrderType):
			if tp == nil {
				consider("tp", trigger, at)
			}
		case IsStopLossType(ord.StopOrderType):
			if sl == nil {
				consider("sl", trigger, at)
			}
		}
	}
	return tp, sl
}

func positionIdxCompatible(episodeIdx, orderIdx int) bool {
	if episodeIdx == models.PositionIdxOneWay || orderIdx == models.PositionIdxOneWay {
		return true
	}
	return episodeIdx == orderIdx
}

func AllocateFillStats(part FillPart) (fee, pnl float64) {
	ratio := 1.0
	if fullSize := part.Fill.Qty(); fullSize > sizeEpsilon {
		ratio = part.Size / fullSize
		if ratio > 1 {
			ratio = 1
		}
	}
	fee = part.Fill.Fee() * ratio
	pnl = part.Fill.Pnl() * ratio
	return fee, pnl
}
