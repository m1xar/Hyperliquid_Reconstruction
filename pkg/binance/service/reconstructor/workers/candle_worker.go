package workers

import (
	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/executors"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/models"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/service/reconstructor/helpers"
	"github.com/m1xar/scope360-reconstruction/pkg/reconstruction/candlespan"
)

func StartCandleWorkers(
	client *resty.Client,
	requests <-chan helpers.CandleRequest,
	workerCount int,
) {
	for i := 0; i < workerCount; i++ {
		go func() {
			for req := range requests {
				candles, err := fetchSpan(client, req)
				req.ReplyCh <- helpers.CandleResponse{Candles: candles, Err: err}
			}
		}()
	}
}

func fetchSpan(client *resty.Client, req helpers.CandleRequest) ([]models.Candle, error) {
	var out []models.Candle

	startMs := executors.AlignToInterval(req.StartMs, candlespan.Minute)
	for _, segment := range candlespan.Split(startMs, req.EndMs) {
		candles, err := executors.FetchCandles(client, req.Symbol, segment.Interval, segment.StartMs, segment.EndMs)
		if err != nil {
			return nil, err
		}
		out = append(out, candles...)
	}
	return out, nil
}
