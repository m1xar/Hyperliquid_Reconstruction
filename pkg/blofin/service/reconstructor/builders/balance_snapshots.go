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

func BuildSyntheticBalanceSnapshots(
	currentEquity float64,
	transfers []models.Transfer,
	positions []domain.Position,
) []domain.UserBalanceSnapshot {
	var events []balanceEvent
	var firstTransferIn *time.Time

	for _, tr := range transfers {
		if !isStableCurrency(tr.Currency) {
			continue
		}

		from := IsTradingAccount(tr.FromAccount)
		to := IsTradingAccount(tr.ToAccount)
		amount := helpers.MustFloat(tr.Amount)
		if amount == 0 {
			continue
		}
		at := helpers.TimeFromMs(tr.Ts)

		switch {
		case to && !from:
			events = append(events, balanceEvent{At: at, Delta: amount})
			if firstTransferIn == nil || at.Before(*firstTransferIn) {
				firstTransferIn = &at
			}
		case from && !to:
			events = append(events, balanceEvent{At: at, Delta: -amount})
		}
	}

	if firstTransferIn == nil {
		return nil
	}

	start := time.Date(
		firstTransferIn.Year(), firstTransferIn.Month(), firstTransferIn.Day(),
		0, 0, 0, 0, time.UTC,
	)

	for _, pos := range positions {
		if pos.ClosedAt == nil || pos.ClosedAt.Before(start) {
			continue
		}
		if !isStableSettledPair(pos.Pair) {
			continue
		}
		events = append(events, balanceEvent{At: *pos.ClosedAt, Delta: pos.NetPnl})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].At.Before(events[j].At)
	})

	snapshots := []domain.UserBalanceSnapshot{
		{CreatedAt: start, Balance: 0},
	}

	balance := 0.0
	for _, ev := range events {
		if ev.At.Before(start) {
			continue
		}
		balance = clampBalance(helpers.Round8(balance + ev.Delta))
		snapshots = append(snapshots, domain.UserBalanceSnapshot{
			CreatedAt: ev.At,
			Balance:   balance,
		})
	}

	snapshots = append(snapshots, domain.UserBalanceSnapshot{
		CreatedAt: time.Now().UTC(),
		Balance:   clampBalance(helpers.Round8(currentEquity)),
	})

	return snapshots
}

func clampBalance(balance float64) float64 {
	if balance < 0 {
		return 0
	}
	return balance
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
