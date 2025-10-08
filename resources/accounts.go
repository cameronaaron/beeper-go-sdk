package resources

import "context"

// Accounts handles account-related API operations
type Accounts struct {
	client ClientInterface
}

// ClientInterface defines the interface for making API requests
type ClientInterface interface {
	DoRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error
	DoRequestWithQuery(ctx context.Context, method, path string, query map[string]interface{}, result interface{}) error
}

// NewAccounts creates a new Accounts resource client
func NewAccounts(client ClientInterface) *Accounts {
	return &Accounts{client: client}
}

// Account represents a chat account added to Beeper
type Account struct {
	AccountID string `json:"accountID"`
	Network   string `json:"network"`
	User      User   `json:"user"`
}

// AccountListResponse represents the response from listing accounts
type AccountListResponse []Account

// List retrieves all connected Beeper accounts available on this device
func (a *Accounts) List(ctx context.Context) (*AccountListResponse, error) {
	var result AccountListResponse
	err := a.client.DoRequest(ctx, "GET", "/v0/get-accounts", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
