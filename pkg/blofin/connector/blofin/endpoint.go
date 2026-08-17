package blofin

import "fmt"

const BaseURL = "https://openapi.blofin.com"

func CheckAccount(apiKey, secret, passphrase string) error {
	client := NewBaseClient()
	AttachAuth(client, Credentials{
		APIKey:     apiKey,
		Secret:     secret,
		Passphrase: passphrase,
	})

	type balance struct {
		TotalEquity string `json:"totalEquity"`
	}
	if _, err := DoGet[balance](client, BaseURL, "/api/v1/account/balance", nil); err != nil {
		return fmt.Errorf("blofin: invalid credentials: %w", err)
	}

	return nil
}
