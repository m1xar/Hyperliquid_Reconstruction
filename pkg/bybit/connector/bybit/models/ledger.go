package models

const (
	LedgerTrade       = "TRADE"
	LedgerSettlement  = "SETTLEMENT"
	LedgerTransferIn  = "TRANSFER_IN"
	LedgerTransferOut = "TRANSFER_OUT"
	LedgerDelivery    = "DELIVERY"
	LedgerLiquidation = "LIQUIDATION"
	LedgerADL         = "ADL"
	LedgerFeeRefund   = "FEE_REFUND"
	LedgerBonus       = "BONUS"

	CategoryLinear = "linear"
)

type LedgerEntry struct {
	ID              string `json:"id"`
	Symbol          string `json:"symbol"`
	Category        string `json:"category"`
	Side            string `json:"side"`
	TransactionTime Flex   `json:"transactionTime"`
	Type            string `json:"type"`
	TransSubType    string `json:"transSubType"`
	Qty             string `json:"qty"`
	Size            string `json:"size"`
	Currency        string `json:"currency"`
	TradePrice      string `json:"tradePrice"`
	Funding         string `json:"funding"`
	Fee             string `json:"fee"`
	CashFlow        string `json:"cashFlow"`
	Change          string `json:"change"`
	CashBalance     string `json:"cashBalance"`
	FeeRate         string `json:"feeRate"`
	BonusChange     string `json:"bonusChange"`
	TradeID         string `json:"tradeId"`
	OrderID         string `json:"orderId"`
	OrderLinkID     string `json:"orderLinkId"`
}

func (e LedgerEntry) IsFill() bool {
	switch e.Type {
	case LedgerTrade, LedgerLiquidation, LedgerADL, LedgerDelivery:
		return e.Category == CategoryLinear && e.Symbol != ""
	}
	return false
}
