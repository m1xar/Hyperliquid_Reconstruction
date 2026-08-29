package helpers

import (
	"sort"
	"strings"

	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
	"github.com/m1xar/scope360-reconstruction/pkg/reconstruction/reverse"
)

func FillKey(fill models.Fill) string {
	return episodeKey(fill)
}

func FillDelta(fill models.Fill, instruments map[string]models.Instrument) float64 {
	return fillSign(fill) * ContractsToBase(fill.FillSize, InstrumentFor(instruments, fill.InstID))
}

func SeedFromOpenPositions(
	positions []models.OpenPosition,
	instruments map[string]models.Instrument,
) map[string]float64 {
	out := make(map[string]float64, len(positions))
	for _, pos := range positions {
		size := ContractsToBase(pos.Positions, InstrumentFor(instruments, pos.InstID))
		if size == 0 {
			continue
		}
		if SideFromOpenPosition(pos.PositionSide, pos.Positions) == "SHORT" {
			size = -abs(size)
		} else {
			size = abs(size)
		}

		side := strings.ToLower(strings.TrimSpace(pos.PositionSide))
		key := pos.InstID
		if side != "" && side != "net" {
			key = pos.InstID + "|" + side
		}
		out[key] += size
	}
	return out
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

type FillSegmenter struct {
	walker      *reverse.Walker[models.Fill]
	instruments map[string]models.Instrument
}

func NewFillSegmenter(
	openPositions []models.OpenPosition,
	instruments map[string]models.Instrument,
) *FillSegmenter {
	seed := reverse.SeedFromMap(SeedFromOpenPositions(openPositions, instruments))
	return &FillSegmenter{
		walker:      reverse.NewWalker[models.Fill](seed),
		instruments: instruments,
	}
}

func (s *FillSegmenter) PushOlderBatch(fills []models.Fill) [][]models.Fill {
	batch := make([]models.Fill, len(fills))
	copy(batch, fills)
	sort.SliceStable(batch, func(i, j int) bool {
		ti, tj := MustInt64(batch[i].Ts), MustInt64(batch[j].Ts)
		if ti == tj {
			return batch[i].TradeID > batch[j].TradeID
		}
		return ti > tj
	})

	groups := make([][]models.Fill, 0)
	for _, fill := range batch {
		delta := FillDelta(fill, s.instruments)
		if delta == 0 {
			continue
		}
		if group, ok := s.walker.Push(FillKey(fill), delta, fill); ok {
			groups = append(groups, group.Fills)
		}
	}
	return groups
}

func (s *FillSegmenter) Flat() bool {
	return s.walker.Flat()
}
