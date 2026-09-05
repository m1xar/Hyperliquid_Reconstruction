package models

import (
	"encoding/json"
	"errors"
	"strconv"
)

type KlineResult struct {
	Category string   `json:"category"`
	Symbol   string   `json:"symbol"`
	List     []Candle `json:"list"`
}

type Candle struct {
	StartTime int64
	O         string
	H         string
	L         string
	C         string
	Vol       string
	Turnover  string
}

func (c *Candle) UnmarshalJSON(data []byte) error {
	var row []string
	if err := json.Unmarshal(data, &row); err != nil {
		return err
	}
	if len(row) < 5 {
		return errors.New("bybit: kline row has insufficient length")
	}
	ts, err := strconv.ParseInt(row[0], 10, 64)
	if err != nil {
		return errors.New("bybit: kline start time invalid")
	}
	c.StartTime = ts
	c.O, c.H, c.L, c.C = row[1], row[2], row[3], row[4]
	if len(row) > 5 {
		c.Vol = row[5]
	}
	if len(row) > 6 {
		c.Turnover = row[6]
	}
	return nil
}
