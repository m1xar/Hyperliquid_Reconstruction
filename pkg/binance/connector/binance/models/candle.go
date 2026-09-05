package models

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Candle struct {
	OpenTime  int64
	O         string
	H         string
	L         string
	C         string
	Vol       string
	CloseTime int64
	Trades    int64
}

func (c *Candle) UnmarshalJSON(data []byte) error {
	var row []any
	if err := json.Unmarshal(data, &row); err != nil {
		return err
	}
	if len(row) < 7 {
		return errors.New("binance: kline row has insufficient length")
	}

	var ok bool
	if c.OpenTime, ok = asInt64(row[0]); !ok {
		return errors.New("binance: kline open time invalid")
	}
	if c.CloseTime, ok = asInt64(row[6]); !ok {
		return errors.New("binance: kline close time invalid")
	}
	if len(row) > 8 {
		c.Trades, _ = asInt64(row[8])
	}

	for i, dst := range []*string{&c.O, &c.H, &c.L, &c.C, &c.Vol} {
		s, ok := row[i+1].(string)
		if !ok {
			return fmt.Errorf("binance: kline field %d invalid", i+1)
		}
		*dst = s
	}
	return nil
}

func asInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case int64:
		return t, true
	case int:
		return int64(t), true
	case json.Number:
		n, err := t.Int64()
		return n, err == nil
	default:
		return 0, false
	}
}
