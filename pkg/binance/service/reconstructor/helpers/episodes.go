package helpers

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/models"
)

const sizeEpsilon = 1e-9

type FillPart struct {
	Fill  models.Trade
	Size  float64
	Price float64
	At    time.Time
	Sign  float64
}

type Episode struct {
	Symbol       string
	PositionSide string
	OpenSign     float64
	Parts        []FillPart
	OpenAt       time.Time
	CloseAt      time.Time
	PeakSize     float64
	Closed       bool
}

func OpenEpisodes(fills []models.Trade) map[string]Episode {
	out := make(map[string]Episode)
	for _, ep := range BuildEpisodes(fills) {
		if ep.Closed {
			continue
		}
		out[EpisodeKey(ep.Symbol, ep.Side())] = ep
	}
	return out
}

func EpisodeKey(symbol, side string) string {
	return strings.ToUpper(symbol) + "|" + strings.ToUpper(side)
}

func BuildEpisodes(fills []models.Trade) []Episode {
	groups := make(map[string][]models.Trade)
	keys := make([]string, 0)
	for _, fill := range fills {
		key := episodeKey(fill)
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], fill)
	}
	sort.Strings(keys)

	episodes := make([]Episode, 0)
	for _, key := range keys {
		episodes = append(episodes, accumulateEpisodes(groups[key])...)
	}

	sort.SliceStable(episodes, func(i, j int) bool {
		return episodes[i].CloseAt.Before(episodes[j].CloseAt)
	})

	return episodes
}

func BuildEpisodesFromGroups(groups [][]models.Trade) []Episode {
	episodes := make([]Episode, 0, len(groups))
	for _, group := range groups {
		episodes = append(episodes, accumulateEpisodes(group)...)
	}

	sort.SliceStable(episodes, func(i, j int) bool {
		return episodes[i].CloseAt.Before(episodes[j].CloseAt)
	})

	return episodes
}

func accumulateEpisodes(group []models.Trade) []Episode {
	sorted := make([]models.Trade, len(group))
	copy(sorted, group)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Time == sorted[j].Time {
			return sorted[i].ID < sorted[j].ID
		}
		return sorted[i].Time < sorted[j].Time
	})

	episodes := make([]Episode, 0, 1)
	var current *Episode
	net := 0.0

	for _, fill := range sorted {
		remaining := MustFloat(fill.Qty)
		if remaining <= sizeEpsilon {
			continue
		}
		at := TimeFromMs(fill.Time)
		sign := fillSign(fill)
		price := MustFloat(fill.Price)

		for remaining > sizeEpsilon {
			if current == nil {
				current = &Episode{
					Symbol:       fill.Symbol,
					PositionSide: strings.ToLower(strings.TrimSpace(fill.PositionSide)),
					OpenSign:     sign,
					OpenAt:       at,
				}
				net = 0
			}

			partSize := remaining
			if math.Abs(net) > sizeEpsilon && sign != current.OpenSign {
				partSize = math.Min(math.Abs(net), remaining)
			}

			current.Parts = append(current.Parts, FillPart{
				Fill:  fill,
				Size:  partSize,
				Price: price,
				At:    at,
				Sign:  sign,
			})

			net += sign * partSize
			if absNet := math.Abs(net); absNet > current.PeakSize {
				current.PeakSize = absNet
			}
			remaining = Round8(remaining - partSize)

			if math.Abs(net) < sizeEpsilon {
				current.CloseAt = at
				current.Closed = true
				episodes = append(episodes, *current)
				current = nil
				net = 0
			}
		}
	}

	if current != nil {
		current.CloseAt = current.Parts[len(current.Parts)-1].At
		episodes = append(episodes, *current)
	}

	return episodes
}

func (e Episode) Side() string {
	if e.OpenSign < 0 {
		return "SHORT"
	}
	return "LONG"
}

func episodeKey(fill models.Trade) string {
	return positionKey(fill.Symbol, fill.PositionSide)
}

func positionKey(symbol, positionSide string) string {
	side := strings.ToLower(strings.TrimSpace(positionSide))
	if side == "" || side == "both" {
		return symbol
	}
	return symbol + "|" + side
}

func fillSign(fill models.Trade) float64 {
	if strings.EqualFold(strings.TrimSpace(fill.Side), "sell") {
		return -1
	}
	return 1
}
