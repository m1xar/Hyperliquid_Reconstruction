package helpers

import (
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/executors"
)

const (
	minuteMs       = int64(60 * 1000)
	feeKlineChunk  = 1000
	feeQuoteAsset  = "USDT"
	feeQuoteSuffix = "USDT"
)

type FeeConverter struct {
	client *resty.Client

	mu     sync.Mutex
	cache  map[string]map[int64]float64
	latest map[string]float64
	failed map[string]bool
}

func NewFeeConverter(client *resty.Client) *FeeConverter {
	return &FeeConverter{
		client: client,
		cache:  make(map[string]map[int64]float64),
		latest: make(map[string]float64),
		failed: make(map[string]bool),
	}
}

func (c *FeeConverter) Rate(asset string, tsMs int64) float64 {
	asset = strings.ToUpper(strings.TrimSpace(asset))
	if asset == "" || executors.IsStableAsset(asset) {
		return 1
	}

	minute := tsMs - tsMs%minuteMs

	c.mu.Lock()
	defer c.mu.Unlock()

	if prices, ok := c.cache[asset]; ok {
		if px, ok := prices[minute]; ok {
			return px
		}
	}

	c.prime(asset, minute)

	if px, ok := c.cache[asset][minute]; ok {
		return px
	}
	if px := c.nearestBefore(asset, minute); px > 0 {
		return px
	}
	return c.latestClose(asset)
}

func (c *FeeConverter) prime(asset string, minute int64) {
	if c.failed[asset] {
		return
	}
	symbol := asset + feeQuoteSuffix
	candles, err := executors.FetchCandles(c.client, symbol, "1m", minute, minute+feeKlineChunk*minuteMs-1)
	if err != nil {
		c.failed[asset] = true
		return
	}
	if c.cache[asset] == nil {
		c.cache[asset] = make(map[int64]float64, len(candles))
	}
	for _, k := range candles {
		c.cache[asset][k.StartTime] = MustFloat(k.C)
	}
}

func (c *FeeConverter) nearestBefore(asset string, minute int64) float64 {
	prices := c.cache[asset]
	if len(prices) == 0 {
		return 0
	}
	best := int64(math.MinInt64)
	for open := range prices {
		if open <= minute && open > best {
			best = open
		}
	}
	if best == math.MinInt64 {
		return 0
	}
	return prices[best]
}

func (c *FeeConverter) latestClose(asset string) float64 {
	if px, ok := c.latest[asset]; ok {
		return px
	}
	k, err := executors.FetchLatestClose(c.client, asset+feeQuoteSuffix)
	if err != nil {
		c.latest[asset] = 0
		return 0
	}
	px := MustFloat(k.C)
	c.latest[asset] = px
	return px
}

func NormalizeFees(client *resty.Client, fills []Fill) {
	converter := NewFeeConverter(client)

	idx := make([]int, 0)
	for i := range fills {
		if !executors.IsStableAsset(fills[i].Entry.Currency) && fills[i].Fee() != 0 {
			idx = append(idx, i)
		}
	}
	if len(idx) == 0 {
		return
	}

	sort.Slice(idx, func(a, b int) bool {
		return fills[idx[a]].TimeMs() < fills[idx[b]].TimeMs()
	})

	for _, i := range idx {
		fill := &fills[i]
		rate := converter.Rate(fill.Entry.Currency, fill.TimeMs())
		fill.Entry.Fee = FormatFloat(Round8(fill.Fee() * rate))
		fill.Entry.CashFlow = FormatFloat(Round8(fill.Pnl() * rate))
		fill.Entry.Currency = feeQuoteAsset
	}
}
