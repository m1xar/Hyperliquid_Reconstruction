package helpers

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
	"github.com/m1xar/scope360-reconstruction/pkg/domain"
)

func MustFloat(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}

func Round8(val float64) float64 {
	return math.Round(val*1e8) / 1e8
}

func FormatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func TimeFromMs(ms int64) time.Time {
	return time.UnixMilli(ms).UTC()
}

func CutoffFromDays(days int) *time.Time {
	if days <= 0 {
		return nil
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	return &cutoff
}

func SideFromPosition(side string) string {
	if strings.EqualFold(strings.TrimSpace(side), "sell") {
		return "SHORT"
	}
	return "LONG"
}

func OrderTypeFromBybit(orderType string) string {
	if strings.EqualFold(strings.TrimSpace(orderType), "limit") {
		return "LIMIT"
	}
	return "MARKET"
}

func OrderSideFromBybit(side string) string {
	switch strings.ToLower(strings.TrimSpace(side)) {
	case "buy":
		return "BUY"
	case "sell":
		return "SELL"
	default:
		return strings.ToUpper(side)
	}
}

func IsTakeProfitType(stopOrderType string) bool {
	switch strings.TrimSpace(stopOrderType) {
	case "TakeProfit", "PartialTakeProfit":
		return true
	}
	return false
}

func IsStopLossType(stopOrderType string) bool {
	switch strings.TrimSpace(stopOrderType) {
	case "StopLoss", "PartialStopLoss", "TrailingStop":
		return true
	}
	return false
}

func GetHighLow(candles []models.Candle) (high, low *float64) {
	if len(candles) == 0 {
		return nil, nil
	}

	h := MustFloat(candles[0].H)
	l := MustFloat(candles[0].L)

	for _, c := range candles {
		if v := MustFloat(c.H); v > h {
			h = v
		}
		if v := MustFloat(c.L); v < l {
			l = v
		}
	}

	return &h, &l
}

func NormalizePair(symbol string) string {
	var b strings.Builder
	for _, r := range symbol {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func BalanceWindowStart(positions []domain.Position, cutoff *time.Time) *time.Time {
	if cutoff == nil {
		return nil
	}

	start := *cutoff
	for _, pos := range positions {
		if pos.CreatedAt.Before(start) {
			start = pos.CreatedAt
		}
	}
	return &start
}
