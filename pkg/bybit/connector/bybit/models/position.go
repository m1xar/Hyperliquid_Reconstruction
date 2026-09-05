package models

const (
	PositionIdxOneWay = 0
	PositionIdxLong   = 1
	PositionIdxShort  = 2
)

type Position struct {
	PositionIdx    Flex   `json:"positionIdx"`
	Symbol         string `json:"symbol"`
	Side           string `json:"side"`
	Size           string `json:"size"`
	AvgPrice       string `json:"avgPrice"`
	PositionValue  string `json:"positionValue"`
	Leverage       string `json:"leverage"`
	MarkPrice      string `json:"markPrice"`
	LiqPrice       string `json:"liqPrice"`
	TakeProfit     string `json:"takeProfit"`
	StopLoss       string `json:"stopLoss"`
	UnrealisedPnl  string `json:"unrealisedPnl"`
	CurRealisedPnl string `json:"curRealisedPnl"`
	CumRealisedPnl string `json:"cumRealisedPnl"`
	PositionStatus string `json:"positionStatus"`
	CreatedTime    Flex   `json:"createdTime"`
	UpdatedTime    Flex   `json:"updatedTime"`
	OpenTime       Flex   `json:"openTime"`
	Seq            Flex   `json:"seq"`
}
