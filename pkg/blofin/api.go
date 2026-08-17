package blofin

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-resty/resty/v2"
	blofinclient "github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/executors"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/service/reconstructor"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/service/reconstructor/builders"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/service/reconstructor/envelope"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/service/reconstructor/helpers"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/service/reconstructor/workers"
	"github.com/m1xar/scope360-reconstruction/pkg/domain"
)

const (
	defaultCandleWorkers   = 4
	defaultPositionWorkers = 8
)

const historyLookback = 1 * time.Minute

func GetAuthStatus(apiKey, secret, passphrase string) string {
	if err := blofinclient.CheckAccount(apiKey, secret, passphrase); err != nil {
		return "error"
	}

	return "ok"
}

func GetBuiltPositions(
	client *resty.Client,
	creds blofinclient.Credentials,
	days int,
) ([]domain.Position, error) {
	blofinclient.AttachAuth(client, creds)

	minCloseMs := int64(0)
	if cutoff := helpers.CutoffFromDays(days); cutoff != nil {
		minCloseMs = cutoff.UnixMilli()
	}
	fills, err := executors.FetchAllFills(client, blofinclient.BaseURL, 0)
	if err != nil {
		return nil, err
	}
	if len(fills) == 0 {
		return []domain.Position{}, nil
	}

	instruments, err := executors.FetchInstruments(client, blofinclient.BaseURL)
	if err != nil {
		return nil, err
	}

	oldestMs := helpers.OldestFillMs(fills) - historyLookback.Milliseconds()

	orders, err := executors.FetchAllOrders(client, blofinclient.BaseURL, oldestMs)
	if err != nil {
		return nil, err
	}

	fundings, err := executors.FetchAllFundingFees(client, blofinclient.BaseURL, oldestMs)
	if err != nil {
		fundings = nil
	}

	candleRequests := make(chan helpers.CandleRequest, defaultCandleWorkers)
	workers.StartCandleWorkers(client, blofinclient.BaseURL, candleRequests, defaultCandleWorkers)

	envelopes := make(chan envelope.PositionEnvelope)
	positionsCh := make(chan domain.Position)

	go func() {
		reconstructor.ReconstructPositions(
			fills, helpers.IndexOrdersByID(orders), fundings, instruments,
			candleRequests, envelopes, minCloseMs,
		)
		close(envelopes)
		close(candleRequests)
	}()

	workers.StartPositionBuilders(envelopes, positionsCh, defaultPositionWorkers)

	positions := make([]domain.Position, 0)
	for pos := range positionsCh {
		positions = append(positions, pos)
	}

	sort.Slice(positions, func(i, j int) bool {
		return positions[i].ClosedAt.Before(*positions[j].ClosedAt)
	})

	if snapshots, err := balanceSnapshots(client, positions); err == nil {
		helpers.AttachBalanceInit(&positions, snapshots)
	}

	return positions, nil
}

func balanceSnapshots(
	client *resty.Client,
	positions []domain.Position,
) ([]domain.UserBalanceSnapshot, error) {
	currentEquity, err := executors.FetchTotalEquity(client, blofinclient.BaseURL)
	if err != nil {
		return nil, err
	}

	transfers, err := executors.FetchAllTransfers(client, blofinclient.BaseURL, 0)
	if err != nil {
		return nil, err
	}

	return builders.BuildSyntheticBalanceSnapshots(currentEquity, transfers, positions), nil
}

func GetClosedPositionByExactMatch(
	client *resty.Client,
	creds blofinclient.Credentials,
	pair string,
	openedAt time.Time,
	side string,
) (*domain.Position, error) {
	positions, err := GetBuiltPositions(client, creds, 0)
	if err != nil {
		return nil, err
	}

	for i := range positions {
		pos := &positions[i]
		if pos.Pair == pair && pos.Side == side && pos.CreatedAt.Equal(openedAt) {
			return pos, nil
		}
	}
	return nil, nil
}

func GetOpenPositions(
	client *resty.Client,
	creds blofinclient.Credentials,
) ([]domain.OpenPosition, error) {
	blofinclient.AttachAuth(client, creds)

	raw, err := executors.FetchOpenPositions(client, blofinclient.BaseURL)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return []domain.OpenPosition{}, nil
	}

	instruments, err := executors.FetchInstruments(client, blofinclient.BaseURL)
	if err != nil {
		return nil, err
	}

	startMs := helpers.OldestPositionMs(raw) - historyLookback.Milliseconds()

	fills, err := executors.FetchAllFills(client, blofinclient.BaseURL, startMs)
	if err != nil {
		return nil, err
	}

	orders, err := executors.FetchAllOrders(client, blofinclient.BaseURL, startMs)
	if err != nil {
		return nil, err
	}

	return builders.BuildOpenPositions(raw, fills, helpers.IndexOrdersByID(orders), instruments), nil
}

func GetBalanceSnapshots(
	client *resty.Client,
	creds blofinclient.Credentials,
	days int,
) ([]domain.UserBalanceSnapshot, error) {

	positions, err := GetBuiltPositions(client, creds, 0)
	if err != nil {
		return nil, err
	}

	snapshots, err := balanceSnapshots(client, positions)
	if err != nil {
		return nil, err
	}

	cutoff := helpers.CutoffFromDays(days)
	if cutoff == nil {
		return snapshots, nil
	}

	filtered := snapshots[:0]
	for _, s := range snapshots {
		if !s.CreatedAt.Before(*cutoff) {
			filtered = append(filtered, s)
		}
	}
	return filtered, nil
}

func GetCurrentBalance(
	client *resty.Client,
	creds blofinclient.Credentials,
) (*float64, error) {
	blofinclient.AttachAuth(client, creds)

	equity, err := executors.FetchTotalEquity(client, blofinclient.BaseURL)
	if err != nil {
		return nil, err
	}

	return &equity, nil
}

func GetTransactions(
	client *resty.Client,
	creds blofinclient.Credentials,
	days int,
) ([]domain.Transaction, error) {
	blofinclient.AttachAuth(client, creds)

	startMs := int64(0)
	cutoff := helpers.CutoffFromDays(days)
	if cutoff != nil {
		startMs = cutoff.UnixMilli()
	}

	transfers, err := executors.FetchAllTransfers(client, blofinclient.BaseURL, startMs)
	if err != nil {
		return nil, err
	}

	transactions := builders.BuildTransactionsFromTransfers(transfers)
	if cutoff == nil {
		return transactions, nil
	}

	filtered := transactions[:0]
	for _, tx := range transactions {
		if !tx.Time.Before(*cutoff) {
			filtered = append(filtered, tx)
		}
	}
	return filtered, nil
}

func GetFundings(
	client *resty.Client,
	creds blofinclient.Credentials,
	days int,
) ([]domain.UserFunding, error) {
	blofinclient.AttachAuth(client, creds)

	startMs := int64(0)
	if cutoff := helpers.CutoffFromDays(days); cutoff != nil {
		startMs = cutoff.UnixMilli()
	}

	fees, err := executors.FetchAllFundingFees(client, blofinclient.BaseURL, startMs)
	if err != nil {
		return nil, err
	}

	return builders.BuildFundings(fees), nil
}

func GetCandles(
	client *resty.Client,
	instID string,
	bar string,
	startTime time.Time,
	endTime time.Time,
) ([]models.Candle, error) {
	if client == nil {
		client = blofinclient.NewBaseClient()
	}

	if endTime.Before(startTime) {
		return nil, fmt.Errorf("endTime must be >= startTime")
	}

	return executors.FetchCandles(
		client, blofinclient.BaseURL, instID, bar,
		startTime.UnixMilli(), endTime.UnixMilli(),
	)
}
