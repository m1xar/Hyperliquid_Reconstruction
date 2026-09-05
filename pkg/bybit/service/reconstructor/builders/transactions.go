package builders

import (
	"math"
	"sort"

	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/executors"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/service/reconstructor/helpers"
	"github.com/m1xar/scope360-reconstruction/pkg/domain"
)

func BuildTransactions(ledger *helpers.Ledger) []domain.Transaction {
	out := make([]domain.Transaction, 0, len(ledger.Transfers))

	for _, tr := range ledger.Transfers {
		if !executors.IsStableAsset(tr.Currency) {
			continue
		}

		amount := math.Abs(helpers.MustFloat(tr.CashFlow))
		if amount == 0 {
			amount = math.Abs(helpers.MustFloat(tr.Change))
		}
		if amount == 0 {
			continue
		}

		typ := domain.TransactionTypeDeposit
		if tr.Type == models.LedgerTransferOut {
			typ = domain.TransactionTypeWithdrawal
		}

		out = append(out, domain.Transaction{
			Time:   helpers.TimeFromMs(tr.TransactionTime.Int64()),
			Type:   typ,
			Amount: helpers.Round8(amount),
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Time.Before(out[j].Time)
	})
	return out
}
