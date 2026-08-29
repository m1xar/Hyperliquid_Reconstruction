package executors

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	kraken "github.com/m1xar/scope360-reconstruction/pkg/kraken/connector/kraken"
	"github.com/m1xar/scope360-reconstruction/pkg/kraken/connector/kraken/models"
)

const positionEventsPath = "/api/history/v3/positions"
const positionEventsRateLimitedTries = 4

func FetchAllPositionEventsSince(client *resty.Client, since time.Time) ([]models.PositionEventElement, error) {
	params := map[string]string{
		"sort": "asc",
	}
	if !since.IsZero() {
		params["since"] = fmt.Sprint(since.UnixMilli())
	}

	var result []models.PositionEventElement
	token := ""
	for {
		if token != "" {
			params["continuation_token"] = token
		} else {
			delete(params, "continuation_token")
		}

		resp, err := kraken.DoGetWithRateLimitRetry[models.PositionEventsResponse](client, positionEventsPath, params, positionEventsRateLimitedTries)
		if err != nil {
			return nil, err
		}
		if len(resp.Elements) == 0 {
			break
		}
		result = append(result, resp.Elements...)
		if resp.ContinuationToken == "" {
			break
		}
		token = resp.ContinuationToken
	}

	return result, nil
}
