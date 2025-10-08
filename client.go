package beeperdesktop

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cameronaaron/beeper-go-sdk/internal"
	"github.com/cameronaaron/beeper-go-sdk/resources"
)

// BeeperDesktop is the main API client for the Beeper Desktop API
type BeeperDesktop struct {
	// Configuration
	accessToken string
	baseURL     string
	timeout     time.Duration
	maxRetries  int
	userAgent   string

	// HTTP client
	httpClient *http.Client
	retryLogic *internal.RetryLogic

	// Resource clients
	Accounts *resources.Accounts
	App      *resources.App
	Chats    *resources.Chats
	Contacts *resources.Contacts
	Messages *resources.Messages
	Token    *resources.Token
}

// New creates a new BeeperDesktop client with the given options
func New(opts ...ClientOption) (*BeeperDesktop, error) {
	config := &ClientConfig{
		AccessToken: os.Getenv("BEEPER_ACCESS_TOKEN"),
		BaseURL:     getEnvWithDefault("BEEPER_DESKTOP_BASE_URL", "http://localhost:23373"),
		Timeout:     30 * time.Second,
		MaxRetries:  2,
		UserAgent:   fmt.Sprintf("beeper-desktop-api-go/%s", Version),
	}

	for _, opt := range opts {
		opt(config)
	}

	if config.AccessToken == "" {
		return nil, &AuthenticationError{
			APIError: APIError{
				Status:  401,
				Message: "access token is required",
			},
		}
	}

	// Ensure base URL ends with /
	if !strings.HasSuffix(config.BaseURL, "/") {
		config.BaseURL += "/"
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: config.Timeout,
		}
	}

	client := &BeeperDesktop{
		accessToken: config.AccessToken,
		baseURL:     config.BaseURL,
		timeout:     config.Timeout,
		maxRetries:  config.MaxRetries,
		userAgent:   config.UserAgent,
		httpClient:  httpClient,
		retryLogic:  internal.NewRetryLogic(config.MaxRetries),
	}

	// Initialize resource clients
	client.Accounts = resources.NewAccounts(client)
	client.App = resources.NewApp(client)
	client.Chats = resources.NewChats(client)
	client.Contacts = resources.NewContacts(client)
	client.Messages = resources.NewMessages(client)
	client.Token = resources.NewToken(client)

	return client, nil
}

// DoRequest performs an HTTP request with retry logic and error handling
func (c *BeeperDesktop) DoRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	return c.retryLogic.Do(ctx, func() error {
		return c.doRequestOnce(ctx, method, path, body, result)
	})
}

// doRequestOnce performs a single HTTP request without retry
func (c *BeeperDesktop) doRequestOnce(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	url := c.baseURL + strings.TrimPrefix(path, "/")

	var reqBody io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = strings.NewReader(string(bodyBytes))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &APIConnectionError{
			BeeperDesktopError: BeeperDesktopError{
				Message: fmt.Sprintf("request failed: %v", err),
			},
			Cause: err,
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp.StatusCode, respBody)
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// DoRequestWithQuery makes an HTTP request with query parameters
func (c *BeeperDesktop) DoRequestWithQuery(ctx context.Context, method, path string, query map[string]interface{}, result interface{}) error {
	// Convert map to url.Values
	queryValues := internal.StructToQueryParams(query)
	if len(queryValues) > 0 {
		path += "?" + queryValues.Encode()
	}
	return c.DoRequest(ctx, method, path, nil, result)
}

// handleErrorResponse converts HTTP error responses to typed errors
func (c *BeeperDesktop) handleErrorResponse(statusCode int, body []byte) error {
	var errorResp struct {
		Error   string            `json:"error"`
		Code    string            `json:"code"`
		Details map[string]string `json:"details"`
	}

	_ = json.Unmarshal(body, &errorResp)

	message := errorResp.Error
	if message == "" {
		message = string(body)
	}

	switch statusCode {
	case 400:
		return &BadRequestError{
			APIError: APIError{
				Status:  statusCode,
				Message: message,
				Code:    errorResp.Code,
				Details: errorResp.Details,
			},
		}
	case 401:
		return &AuthenticationError{
			APIError: APIError{
				Status:  statusCode,
				Message: message,
				Code:    errorResp.Code,
				Details: errorResp.Details,
			},
		}
	case 403:
		return &PermissionDeniedError{
			APIError: APIError{
				Status:  statusCode,
				Message: message,
				Code:    errorResp.Code,
				Details: errorResp.Details,
			},
		}
	case 404:
		return &NotFoundError{
			APIError: APIError{
				Status:  statusCode,
				Message: message,
				Code:    errorResp.Code,
				Details: errorResp.Details,
			},
		}
	case 409:
		return &ConflictError{
			APIError: APIError{
				Status:  statusCode,
				Message: message,
				Code:    errorResp.Code,
				Details: errorResp.Details,
			},
		}
	case 422:
		return &UnprocessableEntityError{
			APIError: APIError{
				Status:  statusCode,
				Message: message,
				Code:    errorResp.Code,
				Details: errorResp.Details,
			},
		}
	case 429:
		return &RateLimitError{
			APIError: APIError{
				Status:  statusCode,
				Message: message,
				Code:    errorResp.Code,
				Details: errorResp.Details,
			},
		}
	default:
		if statusCode >= 500 {
			return &InternalServerError{
				APIError: APIError{
					Status:  statusCode,
					Message: message,
					Code:    errorResp.Code,
					Details: errorResp.Details,
				},
			}
		}
		return &APIError{
			Status:  statusCode,
			Message: message,
			Code:    errorResp.Code,
			Details: errorResp.Details,
		}
	}
}

// getEnvWithDefault returns environment variable value or default
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
