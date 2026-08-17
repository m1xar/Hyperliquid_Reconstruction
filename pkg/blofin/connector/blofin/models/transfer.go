package models

type Transfer struct {
	TransferID  string `json:"transferId"`
	Currency    string `json:"currency"`
	FromAccount string `json:"fromAccount"`
	ToAccount   string `json:"toAccount"`
	Amount      string `json:"amount"`
	Ts          string `json:"ts"`
	ClientID    string `json:"clientId"`
}
