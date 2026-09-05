package models

type ExchangeInfo struct {
	Symbols []Instrument `json:"symbols"`
}

type Instrument struct {
	Symbol            string `json:"symbol"`
	Pair              string `json:"pair"`
	ContractType      string `json:"contractType"`
	Status            string `json:"status"`
	BaseAsset         string `json:"baseAsset"`
	QuoteAsset        string `json:"quoteAsset"`
	MarginAsset       string `json:"marginAsset"`
	PricePrecision    int    `json:"pricePrecision"`
	QuantityPrecision int    `json:"quantityPrecision"`
	OnboardDate       int64  `json:"onboardDate"`
	DeliveryDate      int64  `json:"deliveryDate"`
}
