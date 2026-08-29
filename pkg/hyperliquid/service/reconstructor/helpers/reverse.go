package helpers

import (
	"github.com/m1xar/scope360-reconstruction/pkg/hyperliquid/connector/hyperliquid/models"
	"github.com/m1xar/scope360-reconstruction/pkg/reconstruction/reverse"
)

func FillDelta(f models.RawFill) float64 {
	size := MustFloat(f.Sz)
	if f.Side == "B" {
		return size
	}
	return -size
}

type FillSegmenter struct {
	walker *reverse.Walker[models.RawFill]
}

func NewFillSegmenter() *FillSegmenter {
	return &FillSegmenter{walker: reverse.NewWalker[models.RawFill](nil)}
}

func (s *FillSegmenter) PushOlderBatch(fills []models.RawFill) [][]models.RawFill {
	groups := make([][]models.RawFill, 0)

	for i := len(fills) - 1; i >= 0; i-- {
		f := fills[i]
		if !IsOpen(f.Dir) && !IsClose(f.Dir) {
			continue
		}

		delta := FillDelta(f)
		s.walker.Seed(f.Coin, MustFloat(f.StartPosition)+delta)

		if group, ok := s.walker.Push(f.Coin, delta, f); ok {
			groups = append(groups, group.Fills)
		}
	}

	return groups
}

func (s *FillSegmenter) Flat() bool {
	return s.walker.Flat()
}

func SegmentFills(fills []models.RawFill) [][]models.RawFill {
	return NewFillSegmenter().PushOlderBatch(fills)
}
