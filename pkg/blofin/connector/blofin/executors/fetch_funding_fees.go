package executors

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
)

const fundingFeesPath = "/api/v1/account/funding-fees"

const fundingFeesPageLimit = 1000

const fundingWindowSize = 7 * 24 * time.Hour

const FundingRetention = fundingWindowSize

func FetchAllFundingFees(client *resty.Client, baseURL string, startMs int64) ([]models.FundingFee, error) {
	var result []models.FundingFee

	now := time.Now().UnixMilli()
	earliest := now - FundingRetention.Milliseconds()
	if startMs <= 0 || startMs < earliest {
		startMs = earliest
	}

	windowEnd := now
	for windowEnd > startMs {
		windowBegin := windowEnd - fundingWindowSize.Milliseconds()
		if windowBegin < startMs {
			windowBegin = startMs
		}

		after := ""
		for {
			params := map[string]string{
				"begin": fmt.Sprint(windowBegin),
				"end":   fmt.Sprint(windowEnd),
				"limit": fmt.Sprintf("%d", fundingFeesPageLimit),
			}
			if after != "" {
				params["after"] = after
			}

			page, err := doWithRateLimit(func() ([]models.FundingFee, error) {
				return blofin.DoGet[[]models.FundingFee](client, baseURL, fundingFeesPath, params)
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

			result = append(result, page...)
			if len(page) < fundingFeesPageLimit {
				break
			}
			after = page[len(page)-1].BillID
		}

		windowEnd = windowBegin
	}

	return result, nil
}
