package helpers

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
)

const sizeEpsilon = 1e-9

type FillPart struct {
	Fill  models.Fill
	Size  float64
	Price float64
	At    time.Time
	Sign  float64
}

type Episode struct {
	InstID       string
	PositionSide string
	OpenSign     float64
	Parts        []FillPart
	OpenAt       time.Time
	CloseAt      time.Time
	PeakSize     float64
	Closed       bool
}

func OpenEpisodes(fills []models.Fill, instruments map[string]models.Instrument) map[string]Episode {
	out := make(map[string]Episode)
	for _, ep := range BuildEpisodes(fills, instruments) {
		if ep.Closed {
			continue
		}
		out[EpisodeKey(ep.InstID, ep.Side())] = ep
	}
	return out
}

func EpisodeKey(instID, side string) string {
	return strings.ToUpper(instID) + "|" + strings.ToUpper(side)
}

func BuildEpisodes(fills []models.Fill, instruments map[string]models.Instrument) []Episode {
	groups := make(map[string][]models.Fill)
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
		episodes = append(episodes, accumulateEpisodes(groups[key], instruments)...)
	}

	sort.SliceStable(episodes, func(i, j int) bool {
		return episodes[i].CloseAt.Before(episodes[j].CloseAt)
	})

	return episodes
}

func BuildEpisodesFromGroups(
	groups [][]models.Fill,
	instruments map[string]models.Instrument,
) []Episode {
	episodes := make([]Episode, 0, len(groups))
	for _, group := range groups {
		episodes = append(episodes, accumulateEpisodes(group, instruments)...)
	}

	sort.SliceStable(episodes, func(i, j int) bool {
		return episodes[i].CloseAt.Before(episodes[j].CloseAt)
	})

	return episodes
}

func accumulateEpisodes(group []models.Fill, instruments map[string]models.Instrument) []Episode {
	sorted := make([]models.Fill, len(group))
	copy(sorted, group)
	sort.SliceStable(sorted, func(i, j int) bool {
		ti, tj := MustInt64(sorted[i].Ts), MustInt64(sorted[j].Ts)
		if ti == tj {
			return sorted[i].TradeID < sorted[j].TradeID
		}
		return ti < tj
	})

	episodes := make([]Episode, 0, 1)
	var current *Episode
	net := 0.0

	for _, fill := range sorted {
		instrument := InstrumentFor(instruments, fill.InstID)
		remaining := ContractsToBase(fill.FillSize, instrument)
		if remaining <= sizeEpsilon {
			continue
		}
		at := TimeFromMs(fill.Ts)
		sign := fillSign(fill)
		price := MustFloat(fill.FillPrice)

		for remaining > sizeEpsilon {
			if current == nil {
				current = &Episode{
					InstID:       fill.InstID,
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

func episodeKey(fill models.Fill) string {
	side := strings.ToLower(strings.TrimSpace(fill.PositionSide))
	if side == "" || side == "net" {
		return fill.InstID
	}
	return fill.InstID + "|" + side
}

func fillSign(fill models.Fill) float64 {
	if strings.EqualFold(strings.TrimSpace(fill.Side), "sell") {
		return -1
	}
	return 1
}

func InstrumentFor(instruments map[string]models.Instrument, instID string) models.Instrument {
	if inst, ok := instruments[instID]; ok && inst.ContractValue != "" {
		return inst
	}
	return models.DefaultInstrument
}
