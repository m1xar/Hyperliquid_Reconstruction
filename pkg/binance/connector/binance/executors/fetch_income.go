package executors

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/models"
)

const incomePath = "/fapi/v1/income"

const incomePageLimit = 1000

const IncomeRetention = 365 * 24 * time.Hour

type incomeKey struct {
	incomeType string
	tranID     int64
}

func FetchAllIncome(client *resty.Client, startMs, endMs int64, incomeType string) ([]models.Income, error) {
	now := time.Now().UnixMilli()
	earliest := now - IncomeRetention.Milliseconds()
	if startMs <= 0 || startMs < earliest {
		startMs = earliest
	}
	if endMs <= 0 || endMs > now {
		endMs = now
	}

	var result []models.Income
	seen := make(map[incomeKey]struct{})
	cursor := startMs

	for cursor <= endMs {
		params := map[string]string{
			"startTime": fmt.Sprint(cursor),
			"endTime":   fmt.Sprint(endMs),
			"limit":     fmt.Sprintf("%d", incomePageLimit),
		}
		if incomeType != "" {
			params["incomeType"] = incomeType
		}

		page, err := doWithRateLimit(func() ([]models.Income, error) {
			return binance.DoGet[[]models.Income](client, incomePath, params, 30)
		})
		if err != nil {
			if len(result) > 0 && isHTTP5xx(err) {
				break
			}
			return nil, err
		}
		if len(page) == 0 {
			break
		}

		added := 0
		maxTs := int64(0)
		for _, row := range page {
			if row.Time > maxTs {
				maxTs = row.Time
			}
			key := incomeKey{incomeType: row.IncomeType, tranID: row.TranID}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, row)
			added++
		}

		if len(page) < incomePageLimit || added == 0 || maxTs <= cursor {
			break
		}
		cursor = maxTs
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Time == result[j].Time {
			return result[i].TranID < result[j].TranID
		}
		return result[i].Time < result[j].Time
	})
	return result, nil
}
