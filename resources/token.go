package resources

import "context"

// Token handles token-related API operations
type Token struct {
	client ClientInterface
}

// NewToken creates a new Token resource client
func NewToken(client ClientInterface) *Token {
	return &Token{client: client}
}

// RevokeRequest represents a token revocation request
type RevokeRequest struct {
	Token         string  `json:"token"`
	TokenTypeHint *string `json:"token_type_hint,omitempty"`
}

// UserInfo represents information about the authenticated user/token
type UserInfo struct {
	Iat      int64   `json:"iat"`                 // Issued at timestamp (Unix epoch seconds)
	Scope    string  `json:"scope"`               // Granted scopes
	Sub      string  `json:"sub"`                 // Subject identifier (token ID)
	TokenUse string  `json:"token_use"`           // Token type
	Aud      *string `json:"aud,omitempty"`       // Audience (client ID)
	ClientID *string `json:"client_id,omitempty"` // Client identifier
	Exp      *int64  `json:"exp,omitempty"`       // Expiration timestamp (Unix epoch seconds)
}

// Info returns information about the authenticated user/token
func (t *Token) Info(ctx context.Context) (*UserInfo, error) {
	var result UserInfo
	err := t.client.DoRequest(ctx, "GET", "/oauth/userinfo", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
