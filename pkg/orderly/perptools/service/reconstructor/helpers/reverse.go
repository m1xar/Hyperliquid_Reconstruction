package helpers

import (
	"sort"
	"strings"

	"github.com/m1xar/scope360-reconstruction/pkg/orderly/perptools/connector/orderly/models"
	"github.com/m1xar/scope360-reconstruction/pkg/reconstruction/reverse"
)

func TradeDelta(trade models.OrderlyTrade) float64 {
	if strings.EqualFold(strings.TrimSpace(trade.Side), "sell") {
		return -trade.ExecutedQuantity
	}
	return trade.ExecutedQuantity
}

func SeedFromOpenPositions(positions []models.OrderlyPosition) map[string]float64 {
	out := make(map[string]float64, len(positions))
	for _, pos := range positions {
		if pos.Symbol == "" {
			continue
		}
		out[pos.Symbol] += pos.PositionQty
	}
	return out
}

type TradeSegmenter struct {
	walker *reverse.Walker[models.OrderlyTrade]
}

func NewTradeSegmenter(openPositions []models.OrderlyPosition) *TradeSegmenter {
	seed := reverse.SeedFromMap(SeedFromOpenPositions(openPositions))
	return &TradeSegmenter{walker: reverse.NewWalker[models.OrderlyTrade](seed)}
}

func (s *TradeSegmenter) PushOlderBatch(trades []models.OrderlyTrade) [][]models.OrderlyTrade {
	batch := make([]models.OrderlyTrade, len(trades))
	copy(batch, trades)
	sort.SliceStable(batch, func(i, j int) bool {
		if batch[i].ExecutedTimestamp == batch[j].ExecutedTimestamp {
			return batch[i].ID > batch[j].ID
		}
		return batch[i].ExecutedTimestamp > batch[j].ExecutedTimestamp
	})

	groups := make([][]models.OrderlyTrade, 0)
	for _, trade := range batch {
		if trade.ExecutedQuantity <= 0 {
			continue
		}
		if group, ok := s.walker.Push(trade.Symbol, TradeDelta(trade), trade); ok {
			groups = append(groups, group.Fills)
		}
	}
	return groups
}

func (s *TradeSegmenter) Flat() bool {
	return s.walker.Flat()
}

func GroupSide(group []models.OrderlyTrade) string {
	if len(group) == 0 {
		return ""
	}
	if TradeDelta(group[0]) < 0 {
		return "SHORT"
	}
	return "LONG"
}
