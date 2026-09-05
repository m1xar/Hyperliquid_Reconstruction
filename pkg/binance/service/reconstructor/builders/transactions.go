package builders

import (
	"math"
	"sort"

	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/executors"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/service/reconstructor/helpers"
	"github.com/m1xar/scope360-reconstruction/pkg/domain"
)

func BuildTransactions(ledger *helpers.Ledger) []domain.Transaction {
	out := make([]domain.Transaction, 0, len(ledger.Transfers))

	for _, tr := range ledger.Transfers {
		if !executors.IsStableAsset(tr.Asset) {
			continue
		}

		amount := helpers.MustFloat(tr.Income)
		if amount == 0 {
			continue
		}

		typ := domain.TransactionTypeDeposit
		if amount < 0 {
			typ = domain.TransactionTypeWithdrawal
		}

		out = append(out, domain.Transaction{
			Time:   helpers.TimeFromMs(tr.Time),
			Type:   typ,
			Amount: helpers.Round8(math.Abs(amount)),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Time.Before(out[j].Time)
	})
	return out
}
