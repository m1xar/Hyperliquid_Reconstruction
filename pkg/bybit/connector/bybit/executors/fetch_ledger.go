package executors

import (
	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
)

const (
	transactionLogPath = "/v5/account/transaction-log"
	LedgerPageLimit    = 50
)

type LedgerFilter struct {
	Category string
	Currency string
	Type     string
}

func FetchLedgerWindow(client *resty.Client, w Window, filter LedgerFilter) ([]models.LedgerEntry, error) {
	params := windowParams(w, map[string]string{
		"accountType": accountTypeUnified,
		"category":    filter.Category,
		"currency":    filter.Currency,
		"type":        filter.Type,
	})
	return collectCursor[models.LedgerEntry](client, transactionLogPath, params, LedgerPageLimit)
}

func FetchLedger(client *resty.Client, startMs, endMs int64, filter LedgerFilter) ([]models.LedgerEntry, error) {
	rows, err := ForEachWindow(Windows(startMs, endMs), DefaultWindowWorkers, func(w Window) ([]models.LedgerEntry, error) {
		return FetchLedgerWindow(client, w, filter)
	})
	if err != nil {
		return nil, err
	}
	return DedupeLedger(rows), nil
}

func DedupeLedger(rows []models.LedgerEntry) []models.LedgerEntry {
	seen := make(map[string]struct{}, len(rows))
	out := rows[:0]
	for _, row := range rows {
		key := row.ID + "|" + row.Type + "|" + row.Currency + "|" + row.TradeID + "|" + row.TransactionTime.String() + "|" + row.Change
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, row)
	}
	return out
}
