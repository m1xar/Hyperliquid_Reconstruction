package helpers

import (
	"math"
	"strings"

	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
)

type Fill struct {
	Entry       models.LedgerEntry
	PositionIdx int
	Seq         int
}

func (f Fill) TimeMs() int64   { return f.Entry.TransactionTime.Int64() }
func (f Fill) Symbol() string  { return f.Entry.Symbol }
func (f Fill) Qty() float64    { return math.Abs(MustFloat(f.Entry.Qty)) }
func (f Fill) Price() float64  { return MustFloat(f.Entry.TradePrice) }
func (f Fill) Fee() float64    { return MustFloat(f.Entry.Fee) }
func (f Fill) Pnl() float64    { return MustFloat(f.Entry.CashFlow) }
func (f Fill) OrderID() string { return f.Entry.OrderID }

func (f Fill) ID() string {
	if f.Entry.TradeID != "" {
		return f.Entry.TradeID
	}
	return f.Entry.ID
}

func (f Fill) Sign() float64 {
	if strings.EqualFold(strings.TrimSpace(f.Entry.Side), "sell") {
		return -1
	}
	return 1
}

func (f Fill) Delta() float64 {
	return f.Sign() * f.Qty()
}

func (f Fill) Key() string {
	return PositionKey(f.Symbol(), f.PositionIdx)
}

func PositionKey(symbol string, positionIdx int) string {
	switch positionIdx {
	case models.PositionIdxLong:
		return symbol + "|long"
	case models.PositionIdxShort:
		return symbol + "|short"
	default:
		return symbol
	}
}

func HedgeSymbols(orders []models.Order, positions []models.Position) map[string]bool {
	out := make(map[string]bool)
	for _, o := range orders {
		if idx := int(o.PositionIdx.Int64()); idx == models.PositionIdxLong || idx == models.PositionIdxShort {
			out[o.Symbol] = true
		}
	}
	for _, p := range positions {
		if idx := int(p.PositionIdx.Int64()); idx == models.PositionIdxLong || idx == models.PositionIdxShort {
			out[p.Symbol] = true
		}
	}
	return out
}

func ResolvePositionIdx(entry models.LedgerEntry, ordersByID map[string]models.Order, hedge map[string]bool) int {
	if order, ok := ordersByID[entry.OrderID]; ok {
		return int(order.PositionIdx.Int64())
	}
	if !hedge[entry.Symbol] {
		return models.PositionIdxOneWay
	}
	size := MustFloat(entry.Size)
	switch {
	case size > 0:
		return models.PositionIdxLong
	case size < 0:
		return models.PositionIdxShort
	case strings.EqualFold(strings.TrimSpace(entry.Side), "sell"):
		return models.PositionIdxLong
	default:
		return models.PositionIdxShort
	}
}

func BuildFills(entries []models.LedgerEntry, ordersByID map[string]models.Order, hedge map[string]bool) []Fill {
	fills := make([]Fill, 0, len(entries))
	for i, entry := range entries {
		if !entry.IsFill() || math.Abs(MustFloat(entry.Qty)) == 0 {
			continue
		}
		fills = append(fills, Fill{
			Entry:       entry,
			PositionIdx: ResolvePositionIdx(entry, ordersByID, hedge),
			Seq:         i,
		})
	}
	return fills
}

func SortFillsAsc(fills []Fill) {
	sortFills(fills, func(a, b Fill) bool {
		if a.TimeMs() == b.TimeMs() {
			return a.Seq > b.Seq
		}
		return a.TimeMs() < b.TimeMs()
	})
}

func SortFillsDesc(fills []Fill) {
	sortFills(fills, func(a, b Fill) bool {
		if a.TimeMs() == b.TimeMs() {
			return a.Seq < b.Seq
		}
		return a.TimeMs() > b.TimeMs()
	})
}
