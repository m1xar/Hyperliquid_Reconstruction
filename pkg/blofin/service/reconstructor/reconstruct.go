package reconstructor

import (
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/service/reconstructor/envelope"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/service/reconstructor/helpers"
)

func ReconstructPositions(
	fills []models.Fill,
	ordersByID map[string]models.Order,
	fundings []models.FundingFee,
	instruments map[string]models.Instrument,
	candleRequests chan<- helpers.CandleRequest,
	out chan<- envelope.PositionEnvelope,
	minCloseMs int64,
) {
	type pendingCandle struct {
		env     envelope.PositionEnvelope
		replyCh chan helpers.CandleResponse
	}

	episodes := helpers.BuildEpisodes(fills, instruments)
	pending := make([]pendingCandle, 0, len(episodes))

	for _, ep := range episodes {
		if !ep.Closed {
			continue
		}
		if minCloseMs > 0 && ep.CloseAt.UnixMilli() < minCloseMs {
			continue
		}

		env := buildEnvelope(ep, ordersByID, fundings, instruments)

		if candleRequests == nil {
			out <- env
			continue
		}

		replyCh := make(chan helpers.CandleResponse, 1)
		candleRequests <- helpers.CandleRequest{
			InstID:  ep.InstID,
			Bar:     "1m",
			StartMs: ep.OpenAt.UnixMilli(),
			EndMs:   ep.CloseAt.UnixMilli(),
			ReplyCh: replyCh,
		}
		pending = append(pending, pendingCandle{env: env, replyCh: replyCh})
	}

	for _, p := range pending {
		resp := <-p.replyCh
		if resp.Err == nil {
			p.env.High, p.env.Low = helpers.GetHighLow(resp.Candles)
		}
		out <- p.env
	}
}

func buildEnvelope(
	ep helpers.Episode,
	ordersByID map[string]models.Order,
	fundings []models.FundingFee,
	instruments map[string]models.Instrument,
) envelope.PositionEnvelope {
	orders := helpers.OrdersForEpisode(ep, ordersByID)
	leverage, isolated, tp, sl := helpers.OrderContext(orders)

	return envelope.PositionEnvelope{
		InstID:     ep.InstID,
		Side:       ep.Side(),
		Instrument: helpers.InstrumentFor(instruments, ep.InstID),
		Parts:      ep.Parts,
		Orders:     orders,
		OpenAt:     ep.OpenAt,
		CloseAt:    ep.CloseAt,
		PeakSize:   ep.PeakSize,
		OpenSign:   ep.OpenSign,
		Closed:     ep.Closed,
		Leverage:   leverage,
		Isolated:   isolated,
		StopLoss:   sl,
		TakeProfit: tp,
		Funding:    helpers.FundingForRange(fundings, ep.InstID, ep.OpenAt.UnixMilli(), ep.CloseAt.UnixMilli()),
	}
}
