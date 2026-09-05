package models

const (
	IncomeTransfer         = "TRANSFER"
	IncomeInternalTransfer = "INTERNAL_TRANSFER"
	IncomeRealizedPnl      = "REALIZED_PNL"
	IncomeFundingFee       = "FUNDING_FEE"
	IncomeSpecialFunding   = "SPECIAL_FUNDING_FEE"
	IncomeCommission       = "COMMISSION"
	IncomeInsuranceClear   = "INSURANCE_CLEAR"
)

type Income struct {
	Symbol     string `json:"symbol"`
	IncomeType string `json:"incomeType"`
	Income     string `json:"income"`
	Asset      string `json:"asset"`
	Info       string `json:"info"`
	Time       int64  `json:"time"`
	TranID     int64  `json:"tranId"`
	TradeID    string `json:"tradeId"`
}
