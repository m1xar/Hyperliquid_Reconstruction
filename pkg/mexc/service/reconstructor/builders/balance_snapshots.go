package builders

import (
	"sort"
	"strings"
	"time"

	"github.com/m1xar/scope360-reconstruction/pkg/domain"
	"github.com/m1xar/scope360-reconstruction/pkg/mexc/connector/mexc/models"
	"github.com/m1xar/scope360-reconstruction/pkg/mexc/service/reconstructor/helpers"
)

type balanceEvent struct {
	At    time.Time
	Delta float64
}

func BuildBalanceSnapshots(
	currentEquity float64,
	transfers []models.TransferRecord,
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
	transfers []models.TransferRecord,
	positions []domain.Position,
	windowStart *time.Time,
) []balanceEvent {
	events := make([]balanceEvent, 0, len(transfers)+len(positions))

	within := func(at time.Time) bool {
		return windowStart == nil || !at.Before(*windowStart)
	}

	for _, tr := range transfers {
		if !isSuccessfulTransfer(tr) || !isStableCurrency(tr.Currency) {
			continue
		}

		at := helpers.TimeFromMs(tr.CreateTime)
		if !within(at) {
			continue
		}

		switch strings.ToUpper(strings.TrimSpace(tr.Type)) {
		case "IN":
			events = append(events, balanceEvent{At: at, Delta: tr.Amount})
		case "OUT":
			events = append(events, balanceEvent{At: at, Delta: -tr.Amount})
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

func isSuccessfulTransfer(tr models.TransferRecord) bool {
	return strings.EqualFold(strings.TrimSpace(tr.State), "SUCCESS")
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
