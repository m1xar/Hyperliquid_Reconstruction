package helpers

import (
	"math"
	"strings"

	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
)

func IndexOrdersByID(orders []models.Order) map[string]models.Order {
	idx := make(map[string]models.Order, len(orders))
	for _, o := range orders {
		idx[o.OrderID] = o
	}
	return idx
}

func OrdersForEpisode(ep Episode, ordersByID map[string]models.Order) map[string]models.Order {
	out := make(map[string]models.Order, len(ep.Parts))
	for _, part := range ep.Parts {
		if order, ok := ordersByID[part.Fill.OrderID]; ok {
			out[part.Fill.OrderID] = order
		}
	}
	return out
}

func OrderContext(orders map[string]models.Order) (leverage float64, isolated bool, tp, sl *float64) {
	for _, order := range orders {
		if lev := MustFloat(order.Leverage); lev > leverage {
			leverage = lev
		}
		if strings.EqualFold(strings.TrimSpace(order.MarginMode), "isolated") {
			isolated = true
		}
		if v := MustFloat(order.TpTriggerPx); v > 0 && tp == nil {
			rounded := Round8(v)
			tp = &rounded
		}
		if v := MustFloat(order.SlTriggerPx); v > 0 && sl == nil {
			rounded := Round8(v)
			sl = &rounded
		}
	}
	return leverage, isolated, tp, sl
}

func AllocateFillStats(part FillPart, instrument models.Instrument) (fee, pnl float64) {
	ratio := 1.0
	if fullSize := ContractsToBase(part.Fill.FillSize, instrument); fullSize > sizeEpsilon {
		ratio = part.Size / fullSize
		if ratio > 1 {
			ratio = 1
		}
	}
	fee = math.Abs(MustFloat(part.Fill.Fee)) * ratio
	pnl = MustFloat(part.Fill.FillPnl) * ratio
	return fee, pnl
}

func FundingForRange(fundings []models.FundingFee, instID string, fromMs, toMs int64) float64 {
	var total float64
	for _, f := range fundings {
		if f.InstID != instID {
			continue
		}
		ts := MustInt64(f.FundingTime)
		if ts == 0 {
			ts = MustInt64(f.Ts)
		}
		if ts < fromMs || ts > toMs {
			continue
		}
		total += MustFloat(f.FundingFee)
	}
	return Round8(total)
}
