package resources

import "context"

// App handles app-related API operations
type App struct {
	client ClientInterface
}

// NewApp creates a new App resource client
func NewApp(client ClientInterface) *App {
	return &App{client: client}
}

// AppDownloadAssetParams represents parameters for downloading an asset
type AppDownloadAssetParams struct {
	AssetURL string `json:"assetUrl"`
}

// AppDownloadAssetResponse represents the response from downloading an asset
type AppDownloadAssetResponse struct {
	LocalPath string `json:"localPath"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// AppOpenParams represents parameters for opening the app
type AppOpenParams struct {
	ChatID          *string `json:"chatId,omitempty"`
	MessageID       *string `json:"messageId,omitempty"`
	DraftText       *string `json:"draftText,omitempty"`
	DraftAttachment *string `json:"draftAttachment,omitempty"`
}

// AppOpenResponse represents the response from opening the app
type AppOpenResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// AppSearchParams represents parameters for searching
type AppSearchParams struct {
	Query            string   `json:"query"`
	AccountIDs       []string `json:"accountIDs,omitempty"`
	ChatType         *string  `json:"chatType,omitempty"`
	IncludeMuted     *bool    `json:"includeMuted,omitempty"`
	Limit            *int     `json:"limit,omitempty"`
	MessageLimit     *int     `json:"messageLimit,omitempty"`
	ParticipantLimit *int     `json:"participantLimit,omitempty"`
}

// AppSearchResponse represents the response from searching
type AppSearchResponse struct {
	Chats    []ChatSearchResult    `json:"chats"`
	Messages []MessageSearchResult `json:"messages"`
}

// ChatSearchResult represents a chat in search results
type ChatSearchResult struct {
	Chat         Chat      `json:"chat"`
	Participants []User    `json:"participants"`
	Messages     []Message `json:"messages"`
}

// MessageSearchResult represents a message in search results
type MessageSearchResult struct {
	Message Message `json:"message"`
	Chat    Chat    `json:"chat"`
}

// DownloadAsset downloads an asset from a URL
func (a *App) DownloadAsset(ctx context.Context, params AppDownloadAssetParams) (*AppDownloadAssetResponse, error) {
	var result AppDownloadAssetResponse
	err := a.client.DoRequest(ctx, "POST", "/v0/download-asset", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Open opens Beeper Desktop and optionally navigates to a specific chat
func (a *App) Open(ctx context.Context, params AppOpenParams) (*AppOpenResponse, error) {
	var result AppOpenResponse
	err := a.client.DoRequest(ctx, "POST", "/v0/open-app", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Search searches for chats and messages in one call
func (a *App) Search(ctx context.Context, params AppSearchParams) (*AppSearchResponse, error) {
	var result AppSearchResponse
	err := a.client.DoRequest(ctx, "GET", "/v0/search", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
