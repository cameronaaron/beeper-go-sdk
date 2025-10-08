package beeperdesktop

import (
	"net/http"
	"time"
)

// ClientConfig holds configuration for the BeeperDesktop client
type ClientConfig struct {
	AccessToken string
	BaseURL     string
	Timeout     time.Duration
	MaxRetries  int
	UserAgent   string
	HTTPClient  *http.Client
}

// ClientOption is a function that modifies ClientConfig
type ClientOption func(*ClientConfig)

// WithAccessToken sets the access token for authentication
func WithAccessToken(token string) ClientOption {
	return func(c *ClientConfig) {
		c.AccessToken = token
	}
}

// WithBaseURL sets the base URL for the API
func WithBaseURL(baseURL string) ClientOption {
	return func(c *ClientConfig) {
		c.BaseURL = baseURL
	}
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.Timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(maxRetries int) ClientOption {
	return func(c *ClientConfig) {
		c.MaxRetries = maxRetries
	}
}

// WithUserAgent sets a custom user agent
func WithUserAgent(userAgent string) ClientOption {
	return func(c *ClientConfig) {
		c.UserAgent = userAgent
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *ClientConfig) {
		c.HTTPClient = httpClient
	}
}
