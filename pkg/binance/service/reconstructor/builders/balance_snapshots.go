package builders

import (
	"time"

	"github.com/m1xar/scope360-reconstruction/pkg/binance/service/reconstructor/helpers"
	"github.com/m1xar/scope360-reconstruction/pkg/domain"
)

type balanceEvent struct {
	At    time.Time
	Delta float64
}

func BuildBalanceSnapshots(
	currentBalance float64,
	ledger *helpers.Ledger,
	windowStart *time.Time,
) []domain.UserBalanceSnapshot {
	events := collectBalanceEvents(ledger, windowStart)

	now := time.Now().UTC()
	if len(events) == 0 {
		return []domain.UserBalanceSnapshot{
			{CreatedAt: now, Balance: helpers.Round8(currentBalance)},
		}
	}

	snapshots := make([]domain.UserBalanceSnapshot, len(events)+2)

	running := currentBalance
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
		Balance:   helpers.Round8(currentBalance),
	}

	return dedupeSameInstant(snapshots)
}

func collectBalanceEvents(ledger *helpers.Ledger, windowStart *time.Time) []balanceEvent {
	entries := ledger.StableEntries()
	events := make([]balanceEvent, 0, len(entries))

	for _, row := range entries {
		delta := helpers.MustFloat(row.Income)
		if delta == 0 {
			continue
		}
		at := helpers.TimeFromMs(row.Time)
		if windowStart != nil && at.Before(*windowStart) {
			continue
		}

		if n := len(events); n > 0 && events[n-1].At.Equal(at) {
			events[n-1].Delta += delta
			continue
		}
		events = append(events, balanceEvent{At: at, Delta: delta})
	}

	return events
}

func dedupeSameInstant(snapshots []domain.UserBalanceSnapshot) []domain.UserBalanceSnapshot {
	if len(snapshots) < 2 {
		return snapshots
	}
	out := snapshots[:1]
	for _, s := range snapshots[1:] {
		if s.CreatedAt.Equal(out[len(out)-1].CreatedAt) {
			out[len(out)-1] = s
			continue
		}
		out = append(out, s)
	}
	return out
}
