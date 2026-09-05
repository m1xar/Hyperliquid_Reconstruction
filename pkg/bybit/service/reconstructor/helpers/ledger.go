package helpers

import (
	"sort"

	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/executors"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
)

type Ledger struct {
	Entries   []models.LedgerEntry
	Transfers []models.LedgerEntry
	Fundings  []models.LedgerEntry

	tradeSymbols map[string]struct{}
	fundingBySym map[string][]models.LedgerEntry
	refundBySym  map[string][]models.LedgerEntry
}

func BuildLedger(newestFirst []models.LedgerEntry) *Ledger {
	sorted := make([]models.LedgerEntry, len(newestFirst))
	for i, row := range newestFirst {
		sorted[len(newestFirst)-1-i] = row
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].TransactionTime.Int64() < sorted[j].TransactionTime.Int64()
	})

	l := &Ledger{
		Entries:      sorted,
		tradeSymbols: make(map[string]struct{}),
		fundingBySym: make(map[string][]models.LedgerEntry),
		refundBySym:  make(map[string][]models.LedgerEntry),
	}

	for _, row := range sorted {
		switch row.Type {
		case models.LedgerTransferIn, models.LedgerTransferOut:
			l.Transfers = append(l.Transfers, row)
		case models.LedgerSettlement:
			if row.Category == models.CategoryLinear && row.Symbol != "" {
				l.Fundings = append(l.Fundings, row)
				l.fundingBySym[row.Symbol] = append(l.fundingBySym[row.Symbol], row)
			}
		case models.LedgerFeeRefund:
			if row.Symbol != "" {
				l.refundBySym[row.Symbol] = append(l.refundBySym[row.Symbol], row)
			}
		}
		if row.IsFill() {
			l.tradeSymbols[row.Symbol] = struct{}{}
		}
	}

	return l
}

func (l *Ledger) TradeSymbols() []string {
	out := make([]string, 0, len(l.tradeSymbols))
	for s := range l.tradeSymbols {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func (l *Ledger) FundingForRange(symbol string, fromMs, toMs int64) float64 {
	var total float64
	for _, f := range l.fundingBySym[symbol] {
		ts := f.TransactionTime.Int64()
		if ts < fromMs || ts > toMs {
			continue
		}
		total += MustFloat(f.Funding)
	}
	return Round8(total)
}

func (l *Ledger) FeeRefundForRange(symbol string, fromMs, toMs int64) float64 {
	var total float64
	for _, r := range l.refundBySym[symbol] {
		ts := r.TransactionTime.Int64()
		if ts < fromMs || ts > toMs {
			continue
		}
		total += MustFloat(r.Change)
	}
	return Round8(total)
}

func (l *Ledger) StableEntries() []models.LedgerEntry {
	out := make([]models.LedgerEntry, 0, len(l.Entries))
	for _, row := range l.Entries {
		if executors.IsStableAsset(row.Currency) {
			out = append(out, row)
		}
	}
	return out
}
