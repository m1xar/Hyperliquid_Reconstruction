package helpers

import (
	"math"
	"strings"

	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
	"github.com/m1xar/scope360-reconstruction/pkg/reconstruction/reverse"
)

func SeedFromOpenPositions(positions []models.Position) map[string]float64 {
	out := make(map[string]float64, len(positions))
	for _, pos := range positions {
		size := math.Abs(MustFloat(pos.Size))
		if size == 0 {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(pos.Side), "sell") {
			size = -size
		}
		out[PositionKey(pos.Symbol, int(pos.PositionIdx.Int64()))] += size
	}
	return out
}

type FillSegmenter struct {
	walker   *reverse.Walker[Fill]
	position map[string]float64
}

func NewFillSegmenter(openPositions []models.Position) *FillSegmenter {
	seed := SeedFromOpenPositions(openPositions)
	position := make(map[string]float64, len(seed))
	for k, v := range seed {
		position[k] = v
	}
	return &FillSegmenter{
		walker:   reverse.NewWalker[Fill](reverse.SeedFromMap(seed)),
		position: position,
	}
}

func (s *FillSegmenter) PushOlderBatch(fills []Fill) [][]Fill {
	batch := make([]Fill, len(fills))
	copy(batch, fills)
	SortFillsDesc(batch)

	groups := make([][]Fill, 0)
	for _, fill := range batch {
		delta := fill.Delta()
		if delta == 0 {
			continue
		}
		key := fill.Key()
		before := s.position[key] - delta
		if math.Abs(before) < reverse.DefaultEpsilon {
			before = 0
		}
		s.position[key] = before

		if group, ok := s.walker.Push(key, delta, fill); ok {
			groups = append(groups, group.Fills)
		}
	}
	return groups
}

func (s *FillSegmenter) Flat() bool {
	return s.walker.Flat()
}

func (s *FillSegmenter) Resolved() bool {
	for _, size := range s.position {
		if size != 0 {
			return false
		}
	}
	return s.walker.Flat()
}
