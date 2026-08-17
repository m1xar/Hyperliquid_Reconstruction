package executors

import (
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
)

const transfersPath = "/api/v1/asset/bills"

const transfersPageLimit = 100

func FetchAllTransfers(client *resty.Client, baseURL string, startMs int64) ([]models.Transfer, error) {
	var result []models.Transfer
	seen := make(map[string]struct{})

	after := ""
	for {
		params := map[string]string{
			"limit": fmt.Sprintf("%d", transfersPageLimit),
		}
		if after != "" {
			params["after"] = after
		}

		page, err := doWithRateLimit(func() ([]models.Transfer, error) {
			return blofin.DoGet[[]models.Transfer](client, baseURL, transfersPath, params)
		})
		if err != nil {
			if after != "" && isHTTP5xx(err) {
				break
			}
			return nil, err
		}
		if len(page) == 0 {
			break
		}

		added := 0
		oldestTs := int64(0)
		reachedCutoff := false

		for _, tr := range page {
			ts, _ := strconv.ParseInt(tr.Ts, 10, 64)
			if ts > 0 && (oldestTs == 0 || ts < oldestTs) {
				oldestTs = ts
			}
			if startMs > 0 && ts > 0 && ts < startMs {
				reachedCutoff = true
				continue
			}

			key := tr.TransferID
			if key == "" {
				key = fmt.Sprintf("%s|%s|%s|%s", tr.Ts, tr.Currency, tr.FromAccount, tr.Amount)
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, tr)
			added++
		}

		if reachedCutoff || added == 0 || oldestTs == 0 || len(page) < transfersPageLimit {
			break
		}
		after = fmt.Sprint(oldestTs)
	}

	return result, nil
}
