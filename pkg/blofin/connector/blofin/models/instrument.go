package models

type Instrument struct {
	InstID         string `json:"instId"`
	InstType       string `json:"instType"`
	BaseCurrency   string `json:"baseCurrency"`
	QuoteCurrency  string `json:"quoteCurrency"`
	ContractValue  string `json:"contractValue"`
	ContractType   string `json:"contractType"`
	SettleCurrency string `json:"settleCurrency"`
	MaxLeverage    string `json:"maxLeverage"`
	MinSize        string `json:"minSize"`
	LotSize        string `json:"lotSize"`
	TickSize       string `json:"tickSize"`
	State          string `json:"state"`
}

var DefaultInstrument = Instrument{
	ContractValue: "1",
}
