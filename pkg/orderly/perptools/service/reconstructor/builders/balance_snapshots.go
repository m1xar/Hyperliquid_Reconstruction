package builders

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/m1xar/scope360-reconstruction/pkg/domain"
	"github.com/m1xar/scope360-reconstruction/pkg/orderly/perptools/connector/orderly/models"
	"github.com/m1xar/scope360-reconstruction/pkg/orderly/perptools/service/reconstructor/helpers"
)

type balanceEvent struct {
	At    time.Time
	Delta float64
}

func isCompletedTransfer(ev models.OrderlyAssetHistory) bool {
	return strings.EqualFold(ev.TransStatus, "COMPLETED")
}

func BuildBalanceSnapshots(
	currentEquity float64,
	assetHistory []models.OrderlyAssetHistory,
	positions []domain.Position,
	markPrices map[string]float64,
	windowStart *time.Time,
) ([]domain.UserBalanceSnapshot, error) {
	events, err := collectBalanceEvents(assetHistory, positions, markPrices, windowStart)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if len(events) == 0 {
		return []domain.UserBalanceSnapshot{
			{CreatedAt: now, Balance: helpers.Round8(currentEquity)},
		}, nil
	}

	snapshots := make([]domain.UserBalanceSnapshot, len(events)+2)

	running := currentEquity
	for i := len(events) - 1; i >= 0; i-- {
		snapshots[i+1] = domain.UserBalanceSnapshot{
			CreatedAt: events[i].At,
			Balance:   helpers.Round8(running),
		}
		running -= events[i].Delta
	}

	start := events[0].At.Add(-time.Second)
	if windowStart != nil && windowStart.Before(start) {
		start = windowStart.UTC()
	}
	snapshots[0] = domain.UserBalanceSnapshot{
		CreatedAt: start,
		Balance:   helpers.Round8(running),
	}
	snapshots[len(snapshots)-1] = domain.UserBalanceSnapshot{
		CreatedAt: now,
		Balance:   helpers.Round8(currentEquity),
	}

	return snapshots, nil
}

func collectBalanceEvents(
	assetHistory []models.OrderlyAssetHistory,
	positions []domain.Position,
	markPrices map[string]float64,
	windowStart *time.Time,
) ([]balanceEvent, error) {
	events := make([]balanceEvent, 0, len(assetHistory)+len(positions))

	within := func(at time.Time) bool {
		return windowStart == nil || !at.Before(*windowStart)
	}

	for _, ev := range assetHistory {
		if !isCompletedTransfer(ev) {
			continue
		}

		at := time.UnixMilli(ev.CreatedTime).UTC()
		if !within(at) {
			continue
		}

		price := tokenUSDCPrice(ev.Token, markPrices)
		if price <= 0 {
			return nil, fmt.Errorf("missing USDC mark price for token %q", ev.Token)
		}

		switch strings.ToUpper(strings.TrimSpace(ev.Side)) {
		case "DEPOSIT":
			events = append(events, balanceEvent{At: at, Delta: math.Abs(ev.Amount) * price})
		case "WITHDRAW", "WITHDRAWAL":
			events = append(events, balanceEvent{At: at, Delta: -math.Abs(ev.Amount) * price})
		}
	}

	for _, pos := range positions {
		if !pos.Closed || pos.ClosedAt == nil {
			continue
		}
		at := pos.ClosedAt.UTC()
		if !within(at) {
			continue
		}
		events = append(events, balanceEvent{At: at, Delta: pos.NetPnl})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].At.Before(events[j].At)
	})
	return events, nil
}

func isStableToken(token string) bool {
	return strings.Contains(strings.ToUpper(strings.TrimSpace(token)), "USD")
}

func tokenUSDCPrice(token string, markPrices map[string]float64) float64 {
	token = canonicalAssetToken(token)
	if token == "" {
		return 0
	}
	if isStableToken(token) {
		return 1
	}
	return markPrices[token]
}

func canonicalAssetToken(token string) string {
	token = strings.ToUpper(strings.TrimSpace(token))
	token = strings.ReplaceAll(token, "-", "_")
	token = strings.ReplaceAll(token, "/", "_")
	token = strings.ReplaceAll(token, " ", "_")
	token = strings.Trim(token, "_")
	token = strings.TrimPrefix(token, "PERP_")
	for _, suffix := range []string{"_PERP", "_USDC", "_USDT", "_USD"} {
		token = strings.TrimSuffix(token, suffix)
	}
	return token
}
