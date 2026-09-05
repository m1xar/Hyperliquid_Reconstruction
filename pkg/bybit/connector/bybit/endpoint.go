package bybit

import "fmt"

const (
	UnifiedStatusClassic  = 1
	UnifiedStatusUTA1     = 3
	UnifiedStatusUTA1Pro  = 4
	UnifiedStatusUTA2     = 5
	UnifiedStatusUTA2Pro  = 6
	accountInfoPathString = "/v5/account/info"
)

type accountInfo struct {
	UnifiedMarginStatus int    `json:"unifiedMarginStatus"`
	MarginMode          string `json:"marginMode"`
}

func CheckAccount(apiKey, secret string) error {
	client := NewBaseClient()
	AttachAuth(client, Credentials{APIKey: apiKey, Secret: secret})

	info, err := DoGet[accountInfo](client, accountInfoPathString, nil)
	if err != nil {
		return fmt.Errorf("bybit: invalid credentials: %w", err)
	}
	if info.UnifiedMarginStatus == UnifiedStatusClassic {
		return fmt.Errorf("bybit: classic account is not supported")
	}
	return nil
}
