package models

type FundingFee struct {
	BillID       string `json:"billId"`
	InstID       string `json:"instId"`
	Currency     string `json:"currency"`
	FundingFee   string `json:"fundingFee"`
	FundingTime  string `json:"fundingTime"`
	MarginMode   string `json:"marginMode"`
	PositionSize string `json:"positionSize"`
	Ts           string `json:"ts"`
}
