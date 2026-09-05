package binance

import "fmt"

func CheckAccount(apiKey, secret string) error {
	client := NewBaseClient()
	AttachAuth(client, Credentials{APIKey: apiKey, Secret: secret})

	type account struct {
		TotalWalletBalance string `json:"totalWalletBalance"`
	}
	if _, err := DoGet[account](client, "/fapi/v3/account", nil, 5); err != nil {
		return fmt.Errorf("binance: invalid credentials: %w", err)
	}
	return nil
}
