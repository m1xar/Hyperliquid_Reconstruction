package models

import "encoding/json"

type Candle struct {
	Ts               string
	O                string
	H                string
	L                string
	C                string
	Vol              string
	VolCurrency      string
	VolCurrencyQuote string
	Confirm          string
}

func (c *Candle) UnmarshalJSON(data []byte) error {
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) >= 9 {
		c.Ts = arr[0]
		c.O = arr[1]
		c.H = arr[2]
		c.L = arr[3]
		c.C = arr[4]
		c.Vol = arr[5]
		c.VolCurrency = arr[6]
		c.VolCurrencyQuote = arr[7]
		c.Confirm = arr[8]
	}
	return nil
}
