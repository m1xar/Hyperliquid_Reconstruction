package models

type OpenPosition struct {
	PositionID         string `json:"positionId"`
	InstID             string `json:"instId"`
	InstType           string `json:"instType"`
	MarginMode         string `json:"marginMode"`
	PositionSide       string `json:"positionSide"`
	Leverage           string `json:"leverage"`
	Positions          string `json:"positions"`
	AvailablePositions string `json:"availablePositions"`
	AveragePrice       string `json:"averagePrice"`
	MarkPrice          string `json:"markPrice"`
	MarginRatio        string `json:"marginRatio"`
	LiquidationPrice   string `json:"liquidationPrice"`
	UnrealizedPnl      string `json:"unrealizedPnl"`
	InitialMargin      string `json:"initialMargin"`
	MaintenanceMargin  string `json:"maintenanceMargin"`
	RealizedPnl        string `json:"realizedPnl"`
	CreateTime         string `json:"createTime"`
	UpdateTime         string `json:"updateTime"`
}
