package resources

import "context"

// Contacts handles contact-related API operations
type Contacts struct {
	client ClientInterface
}

// NewContacts creates a new Contacts resource client
func NewContacts(client ClientInterface) *Contacts {
	return &Contacts{client: client}
}

// ContactSearchParams represents parameters for searching contacts
type ContactSearchParams struct {
	AccountID string `json:"accountID"`
	Query     string `json:"query"`
}

// ContactSearchResponse represents the response from searching contacts
type ContactSearchResponse struct {
	Items []User `json:"items"`
}

// Search searches for contacts/users
func (c *Contacts) Search(ctx context.Context, params ContactSearchParams) (*ContactSearchResponse, error) {
	var result ContactSearchResponse
	queryParams := map[string]interface{}{
		"accountID": params.AccountID,
		"query":     params.Query,
	}
	err := c.client.DoRequestWithQuery(ctx, "GET", "/v0/search-users", queryParams, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
