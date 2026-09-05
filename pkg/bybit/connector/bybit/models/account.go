package models

const (
	MarginModeIsolated  = "ISOLATED_MARGIN"
	MarginModeRegular   = "REGULAR_MARGIN"
	MarginModePortfolio = "PORTFOLIO_MARGIN"
)

type AccountInfo struct {
	UnifiedMarginStatus int    `json:"unifiedMarginStatus"`
	MarginMode          string `json:"marginMode"`
	IsMasterTrader      bool   `json:"isMasterTrader"`
	SpotHedgingStatus   string `json:"spotHedgingStatus"`
	UpdatedTime         string `json:"updatedTime"`
}

type WalletBalance struct {
	List []WalletAccount `json:"list"`
}

type WalletAccount struct {
	AccountType           string       `json:"accountType"`
	TotalEquity           string       `json:"totalEquity"`
	TotalWalletBalance    string       `json:"totalWalletBalance"`
	TotalMarginBalance    string       `json:"totalMarginBalance"`
	TotalAvailableBalance string       `json:"totalAvailableBalance"`
	TotalPerpUPL          string       `json:"totalPerpUPL"`
	TotalInitialMargin    string       `json:"totalInitialMargin"`
	Coin                  []WalletCoin `json:"coin"`
}

type WalletCoin struct {
	Coin            string `json:"coin"`
	Equity          string `json:"equity"`
	UsdValue        string `json:"usdValue"`
	WalletBalance   string `json:"walletBalance"`
	UnrealisedPnl   string `json:"unrealisedPnl"`
	CumRealisedPnl  string `json:"cumRealisedPnl"`
	SpotBorrow      string `json:"spotBorrow"`
	TotalPositionIM string `json:"totalPositionIM"`
	Bonus           string `json:"bonus"`
}
