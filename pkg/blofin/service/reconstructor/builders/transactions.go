package builders

import (
	"math"
	"sort"
	"strings"

	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/service/reconstructor/helpers"
	"github.com/m1xar/scope360-reconstruction/pkg/domain"
)

func IsTradingAccount(account string) bool {
	name := strings.ToLower(strings.TrimSpace(account))
	return name == "futures" || strings.HasSuffix(name, "_contract")
}

func BuildTransactionsFromTransfers(transfers []models.Transfer) []domain.Transaction {
	out := make([]domain.Transaction, 0, len(transfers))

	for _, tr := range transfers {
		from := IsTradingAccount(tr.FromAccount)
		to := IsTradingAccount(tr.ToAccount)

		var typ string
		switch {
		case to && !from:
			typ = domain.TransactionTypeDeposit
		case from && !to:
			typ = domain.TransactionTypeWithdrawal
		default:
			continue
		}

		amount := helpers.Round8(math.Abs(helpers.MustFloat(tr.Amount)))
		if amount == 0 {
			continue
		}

		out = append(out, domain.Transaction{
			Time:   helpers.TimeFromMs(tr.Ts),
			Type:   typ,
			Amount: amount,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Time.Before(out[j].Time)
	})
	return out
}
