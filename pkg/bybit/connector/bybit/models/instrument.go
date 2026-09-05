package models

type Instrument struct {
	Symbol         string         `json:"symbol"`
	ContractType   string         `json:"contractType"`
	Status         string         `json:"status"`
	BaseCoin       string         `json:"baseCoin"`
	QuoteCoin      string         `json:"quoteCoin"`
	SettleCoin     string         `json:"settleCoin"`
	LaunchTime     Flex           `json:"launchTime"`
	DeliveryTime   Flex           `json:"deliveryTime"`
	LotSizeFilter  LotSizeFilter  `json:"lotSizeFilter"`
	LeverageFilter LeverageFilter `json:"leverageFilter"`
}

type LotSizeFilter struct {
	MaxOrderQty string `json:"maxOrderQty"`
	MinOrderQty string `json:"minOrderQty"`
	QtyStep     string `json:"qtyStep"`
}

type LeverageFilter struct {
	MinLeverage  string `json:"minLeverage"`
	MaxLeverage  string `json:"maxLeverage"`
	LeverageStep string `json:"leverageStep"`
}
