package models

type Balance struct {
	Ts             string          `json:"ts"`
	TotalEquity    string          `json:"totalEquity"`
	IsolatedEquity string          `json:"isolatedEquity"`
	Details        []BalanceDetail `json:"details"`
}

type BalanceDetail struct {
	Currency        string `json:"currency"`
	Equity          string `json:"equity"`
	Balance         string `json:"balance"`
	Available       string `json:"available"`
	AvailableEquity string `json:"availableEquity"`
	Frozen          string `json:"frozen"`
	EquityUsd       string `json:"equityUsd"`
	Ts              string `json:"ts"`
}
