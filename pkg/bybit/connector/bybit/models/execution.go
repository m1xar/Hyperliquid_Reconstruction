package models

type Execution struct {
	Symbol        string `json:"symbol"`
	OrderID       string `json:"orderId"`
	OrderLinkID   string `json:"orderLinkId"`
	Side          string `json:"side"`
	OrderPrice    string `json:"orderPrice"`
	OrderQty      string `json:"orderQty"`
	LeavesQty     string `json:"leavesQty"`
	CreateType    string `json:"createType"`
	OrderType     string `json:"orderType"`
	StopOrderType string `json:"stopOrderType"`
	ExecFee       string `json:"execFee"`
	ExecID        string `json:"execId"`
	ExecPrice     string `json:"execPrice"`
	ExecQty       string `json:"execQty"`
	ExecType      string `json:"execType"`
	ExecValue     string `json:"execValue"`
	ExecTime      Flex   `json:"execTime"`
	FeeCurrency   string `json:"feeCurrency"`
	IsMaker       bool   `json:"isMaker"`
	FeeRate       string `json:"feeRate"`
	MarkPrice     string `json:"markPrice"`
	ClosedSize    string `json:"closedSize"`
	Seq           Flex   `json:"seq"`
}
