package models

type Account struct {
	TotalWalletBalance      string         `json:"totalWalletBalance"`
	TotalMarginBalance      string         `json:"totalMarginBalance"`
	TotalUnrealizedProfit   string         `json:"totalUnrealizedProfit"`
	TotalCrossWalletBalance string         `json:"totalCrossWalletBalance"`
	AvailableBalance        string         `json:"availableBalance"`
	Assets                  []AccountAsset `json:"assets"`
}

type AccountAsset struct {
	Asset            string `json:"asset"`
	WalletBalance    string `json:"walletBalance"`
	UnrealizedProfit string `json:"unrealizedProfit"`
	MarginBalance    string `json:"marginBalance"`
	AvailableBalance string `json:"availableBalance"`
	UpdateTime       int64  `json:"updateTime"`
}
