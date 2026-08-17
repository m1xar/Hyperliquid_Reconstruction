package helpers

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
)

func MustFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func MustInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func Round8(val float64) float64 {
	return math.Round(val*1e8) / 1e8
}

func ContractsToBase(size string, instrument models.Instrument) float64 {
	ctVal := MustFloat(instrument.ContractValue)
	if ctVal == 0 {
		ctVal = 1
	}
	return MustFloat(size) * ctVal
}

func SideFromOpenPosition(positionSide, positions string) string {
	switch strings.ToLower(strings.TrimSpace(positionSide)) {
	case "long":
		return "LONG"
	case "short":
		return "SHORT"
	}

	if MustFloat(positions) < 0 {
		return "SHORT"
	}
	return "LONG"
}

func OrderTypeFromBlofin(orderType string) string {
	switch strings.ToLower(strings.TrimSpace(orderType)) {
	case "limit", "post_only", "fok", "ioc":
		return "LIMIT"
	default:
		return "MARKET"
	}
}

func OrderSideFromBlofin(side string) string {
	switch strings.ToLower(strings.TrimSpace(side)) {
	case "buy":
		return "BUY"
	case "sell":
		return "SELL"
	default:
		return strings.ToUpper(side)
	}
}

func TimeFromMs(ms string) time.Time {
	return time.UnixMilli(MustInt64(ms)).UTC()
}

func TimeFromMs64(ms int64) time.Time {
	return time.UnixMilli(ms).UTC()
}

func CutoffFromDays(days int) *time.Time {
	if days <= 0 {
		return nil
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	return &cutoff
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

func IsFilled(order models.Order) bool {
	return MustFloat(order.FilledSize) > 0
}

func NormalizePair(instID string) string {
	var b strings.Builder
	for _, r := range instID {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func OldestFillMs(fills []models.Fill) int64 {
	oldest := int64(0)
	for _, f := range fills {
		ts := MustInt64(f.Ts)
		if ts == 0 {
			continue
		}
		if oldest == 0 || ts < oldest {
			oldest = ts
		}
	}
	return oldest
}

func OldestPositionMs(positions []models.OpenPosition) int64 {
	oldest := int64(0)
	for _, p := range positions {
		ts := MustInt64(p.CreateTime)
		if ts == 0 {
			continue
		}
		if oldest == 0 || ts < oldest {
			oldest = ts
		}
	}
	return oldest
}
