package models

type Fill struct {
	InstID       string `json:"instId"`
	TradeID      string `json:"tradeId"`
	OrderID      string `json:"orderId"`
	FillPrice    string `json:"fillPrice"`
	FillSize     string `json:"fillSize"`
	FillPnl      string `json:"fillPnl"`
	PositionSide string `json:"positionSide"`
	Side         string `json:"side"`
	Fee          string `json:"fee"`
	Ts           string `json:"ts"`
}
