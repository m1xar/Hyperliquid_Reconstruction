package models

type Order struct {
	OrderID        string `json:"orderId"`
	OrderLinkID    string `json:"orderLinkId"`
	Symbol         string `json:"symbol"`
	Price          string `json:"price"`
	Qty            string `json:"qty"`
	Side           string `json:"side"`
	PositionIdx    Flex   `json:"positionIdx"`
	OrderStatus    string `json:"orderStatus"`
	CreateType     string `json:"createType"`
	CancelType     string `json:"cancelType"`
	AvgPrice       string `json:"avgPrice"`
	CumExecQty     string `json:"cumExecQty"`
	CumExecValue   string `json:"cumExecValue"`
	CumExecFee     string `json:"cumExecFee"`
	TimeInForce    string `json:"timeInForce"`
	OrderType      string `json:"orderType"`
	StopOrderType  string `json:"stopOrderType"`
	TriggerPrice   string `json:"triggerPrice"`
	TakeProfit     string `json:"takeProfit"`
	StopLoss       string `json:"stopLoss"`
	TpslMode       string `json:"tpslMode"`
	TpLimitPrice   string `json:"tpLimitPrice"`
	SlLimitPrice   string `json:"slLimitPrice"`
	TriggerBy      string `json:"triggerBy"`
	ReduceOnly     bool   `json:"reduceOnly"`
	CloseOnTrigger bool   `json:"closeOnTrigger"`
	CreatedTime    Flex   `json:"createdTime"`
	UpdatedTime    Flex   `json:"updatedTime"`
}
