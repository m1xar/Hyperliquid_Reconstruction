package builders

import (
	"sort"
	"strings"
	"time"

	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/service/reconstructor/helpers"
	"github.com/m1xar/scope360-reconstruction/pkg/domain"
)

type balanceEvent struct {
	At    time.Time
	Delta float64
}

func BuildBalanceSnapshots(
	currentEquity float64,
	transfers []models.Transfer,
	positions []domain.Position,
	windowStart *time.Time,
) []domain.UserBalanceSnapshot {
	events := collectBalanceEvents(transfers, positions, windowStart)

	now := time.Now().UTC()
	if len(events) == 0 {
		return []domain.UserBalanceSnapshot{
			{CreatedAt: now, Balance: helpers.Round8(currentEquity)},
		}
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

	return snapshots
}

func collectBalanceEvents(
	transfers []models.Transfer,
	positions []domain.Position,
	windowStart *time.Time,
) []balanceEvent {
	events := make([]balanceEvent, 0, len(transfers)+len(positions))

	within := func(at time.Time) bool {
		return windowStart == nil || !at.Before(*windowStart)
	}

	for _, tr := range transfers {
		if !isStableCurrency(tr.Currency) {
			continue
		}

		amount := helpers.MustFloat(tr.Amount)
		if amount == 0 {
			continue
		}

		from := IsTradingAccount(tr.FromAccount)
		to := IsTradingAccount(tr.ToAccount)
		at := helpers.TimeFromMs(tr.Ts)
		if !within(at) {
			continue
		}

		switch {
		case to && !from:
			events = append(events, balanceEvent{At: at, Delta: amount})
		case from && !to:
			events = append(events, balanceEvent{At: at, Delta: -amount})
		}
	}

	for _, pos := range positions {
		if pos.ClosedAt == nil || !isStableSettledPair(pos.Pair) {
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
	return events
}

func isStableCurrency(currency string) bool {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "USDT", "USDC":
		return true
	default:
		return false
	}
}

func isStableSettledPair(pair string) bool {
	p := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(pair), "_", ""))
	return strings.HasSuffix(p, "USDT") || strings.HasSuffix(p, "USDC")
}
