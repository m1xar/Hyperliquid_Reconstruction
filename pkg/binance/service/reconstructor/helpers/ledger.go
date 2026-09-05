package helpers

import (
	"sort"

	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/executors"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/models"
)

type Ledger struct {
	Entries   []models.Income
	Transfers []models.Income
	Fundings  []models.Income

	tradeSymbols   map[string]struct{}
	fundingBySym   map[string][]models.Income
	insuranceBySym map[string][]models.Income
}

func BuildLedger(entries []models.Income) *Ledger {
	sorted := make([]models.Income, len(entries))
	copy(sorted, entries)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Time == sorted[j].Time {
			return sorted[i].TranID < sorted[j].TranID
		}
		return sorted[i].Time < sorted[j].Time
	})

	l := &Ledger{
		Entries:        sorted,
		tradeSymbols:   make(map[string]struct{}),
		fundingBySym:   make(map[string][]models.Income),
		insuranceBySym: make(map[string][]models.Income),
	}

	for _, row := range sorted {
		switch row.IncomeType {
		case models.IncomeTransfer, models.IncomeInternalTransfer:
			l.Transfers = append(l.Transfers, row)
		case models.IncomeFundingFee, models.IncomeSpecialFunding:
			l.Fundings = append(l.Fundings, row)
			l.fundingBySym[row.Symbol] = append(l.fundingBySym[row.Symbol], row)
		case models.IncomeInsuranceClear:
			l.insuranceBySym[row.Symbol] = append(l.insuranceBySym[row.Symbol], row)
			if row.Symbol != "" {
				l.tradeSymbols[row.Symbol] = struct{}{}
			}
		case models.IncomeRealizedPnl, models.IncomeCommission:
			if row.Symbol != "" {
				l.tradeSymbols[row.Symbol] = struct{}{}
			}
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
		if f.Time < fromMs || f.Time > toMs {
			continue
		}
		total += MustFloat(f.Income)
	}
	return Round8(total)
}

func (l *Ledger) InsuranceForRange(symbol string, fromMs, toMs int64) float64 {
	var total float64
	for _, f := range l.insuranceBySym[symbol] {
		if f.Time < fromMs || f.Time > toMs {
			continue
		}
		total -= MustFloat(f.Income)
	}
	return Round8(total)
}

func (l *Ledger) StableEntries() []models.Income {
	out := make([]models.Income, 0, len(l.Entries))
	for _, row := range l.Entries {
		if executors.IsStableAsset(row.Asset) {
			out = append(out, row)
		}
	}
	return out
}
