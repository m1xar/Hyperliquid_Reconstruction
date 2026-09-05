package models

type Order struct {
	OrderID       int64  `json:"orderId"`
	Symbol        string `json:"symbol"`
	ClientOrderID string `json:"clientOrderId"`
	Status        string `json:"status"`
	Price         string `json:"price"`
	AvgPrice      string `json:"avgPrice"`
	OrigQty       string `json:"origQty"`
	ExecutedQty   string `json:"executedQty"`
	CumQuote      string `json:"cumQuote"`
	TimeInForce   string `json:"timeInForce"`
	Type          string `json:"type"`
	OrigType      string `json:"origType"`
	ReduceOnly    bool   `json:"reduceOnly"`
	ClosePosition bool   `json:"closePosition"`
	Side          string `json:"side"`
	PositionSide  string `json:"positionSide"`
	StopPrice     string `json:"stopPrice"`
	WorkingType   string `json:"workingType"`
	PriceProtect  bool   `json:"priceProtect"`
	Time          int64  `json:"time"`
	UpdateTime    int64  `json:"updateTime"`
}
