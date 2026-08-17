package models

type Order struct {
	OrderID       string `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	InstID        string `json:"instId"`
	MarginMode    string `json:"marginMode"`
	PositionSide  string `json:"positionSide"`
	Side          string `json:"side"`
	OrderType     string `json:"orderType"`
	Price         string `json:"price"`
	Size          string `json:"size"`
	ReduceOnly    string `json:"reduceOnly"`
	Leverage      string `json:"leverage"`
	State         string `json:"state"`
	FilledSize    string `json:"filledSize"`
	Pnl           string `json:"pnl"`
	AveragePrice  string `json:"averagePrice"`
	Fee           string `json:"fee"`
	CreateTime    string `json:"createTime"`
	UpdateTime    string `json:"updateTime"`
	OrderCategory string `json:"orderCategory"`
	TpTriggerPx   string `json:"tpTriggerPrice"`
	TpOrderPx     string `json:"tpOrderPrice"`
	SlTriggerPx   string `json:"slTriggerPrice"`
	SlOrderPx     string `json:"slOrderPrice"`
	AlgoID        string `json:"algoId"`
}
