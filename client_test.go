package beeperdesktop

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("with access token", func(t *testing.T) {
		client, err := New(WithAccessToken("test-token"))
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "test-token", client.accessToken)
	})

	t.Run("without access token", func(t *testing.T) {
		client, err := New()
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.IsType(t, &AuthenticationError{}, err)
	})

	t.Run("with custom options", func(t *testing.T) {
		customClient := &http.Client{Timeout: 10 * time.Second}
		client, err := New(
			WithAccessToken("test-token"),
			WithBaseURL("https://api.example.com"),
			WithTimeout(5*time.Second),
			WithMaxRetries(5),
			WithUserAgent("test-agent"),
			WithHTTPClient(customClient),
		)
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com/", client.baseURL)
		assert.Equal(t, 5*time.Second, client.timeout)
		assert.Equal(t, 5, client.maxRetries)
		assert.Equal(t, "test-agent", client.userAgent)
		assert.Equal(t, customClient, client.httpClient)
	})
}

func TestBeeperDesktop_DoRequest(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		}))
		defer server.Close()

		client, err := New(
			WithAccessToken("test-token"),
			WithBaseURL(server.URL),
			WithMaxRetries(0), // Disable retries for testing
		)
		require.NoError(t, err)

		var result map[string]interface{}
		err = client.DoRequest(context.Background(), "GET", "/test", nil, &result)
		require.NoError(t, err)
		assert.Equal(t, true, result["success"])
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "not found", "code": "NOT_FOUND"}`))
		}))
		defer server.Close()

		client, err := New(
			WithAccessToken("test-token"),
			WithBaseURL(server.URL),
			WithMaxRetries(0),
		)
		require.NoError(t, err)

		var result map[string]interface{}
		err = client.DoRequest(context.Background(), "GET", "/test", nil, &result)
		require.Error(t, err)

		notFoundErr, ok := err.(*NotFoundError)
		require.True(t, ok)
		assert.Equal(t, 404, notFoundErr.Status)
		assert.Equal(t, "not found", notFoundErr.Message)
		assert.Equal(t, "NOT_FOUND", notFoundErr.Code)
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, err := New(
			WithAccessToken("test-token"),
			WithBaseURL(server.URL),
			WithMaxRetries(0),
		)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		var result map[string]interface{}
		err = client.DoRequest(ctx, "GET", "/test", nil, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errorType  interface{}
	}{
		{"BadRequest", 400, &BadRequestError{}},
		{"Unauthorized", 401, &AuthenticationError{}},
		{"Forbidden", 403, &PermissionDeniedError{}},
		{"NotFound", 404, &NotFoundError{}},
		{"Conflict", 409, &ConflictError{}},
		{"UnprocessableEntity", 422, &UnprocessableEntityError{}},
		{"TooManyRequests", 429, &RateLimitError{}},
		{"InternalServerError", 500, &InternalServerError{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"error": "test error"}`))
			}))
			defer server.Close()

			client, err := New(
				WithAccessToken("test-token"),
				WithBaseURL(server.URL),
				WithMaxRetries(0),
			)
			require.NoError(t, err)

			var result map[string]interface{}
			err = client.DoRequest(context.Background(), "GET", "/test", nil, &result)
			require.Error(t, err)
			assert.IsType(t, tt.errorType, err)
		})
	}
}
