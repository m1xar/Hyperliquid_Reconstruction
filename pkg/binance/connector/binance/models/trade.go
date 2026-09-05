package models

type Trade struct {
	Symbol          string `json:"symbol"`
	ID              int64  `json:"id"`
	OrderID         int64  `json:"orderId"`
	Side            string `json:"side"`
	PositionSide    string `json:"positionSide"`
	Price           string `json:"price"`
	Qty             string `json:"qty"`
	QuoteQty        string `json:"quoteQty"`
	RealizedPnl     string `json:"realizedPnl"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	MarginAsset     string `json:"marginAsset"`
	Buyer           bool   `json:"buyer"`
	Maker           bool   `json:"maker"`
	Time            int64  `json:"time"`
}
