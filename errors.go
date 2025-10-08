package beeperdesktop

import "fmt"

// BeeperDesktopError is the base error type for all Beeper Desktop API errors
type BeeperDesktopError struct {
	Message string
}

func (e *BeeperDesktopError) Error() string {
	return e.Message
}

// APIError represents an error response from the API
type APIError struct {
	Status  int
	Message string
	Code    string
	Details map[string]string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("API error %d (%s): %s", e.Status, e.Code, e.Message)
	}
	return fmt.Sprintf("API error %d: %s", e.Status, e.Message)
}

// APIConnectionError represents a connection error
type APIConnectionError struct {
	BeeperDesktopError
	Cause error
}

func (e *APIConnectionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("connection error: %s (caused by: %v)", e.Message, e.Cause)
	}
	return fmt.Sprintf("connection error: %s", e.Message)
}

func (e *APIConnectionError) Unwrap() error {
	return e.Cause
}

// APIConnectionTimeoutError represents a timeout error
type APIConnectionTimeoutError struct {
	APIConnectionError
}

// BadRequestError represents a 400 error
type BadRequestError struct {
	APIError
}

// AuthenticationError represents a 401 error
type AuthenticationError struct {
	APIError
}

// PermissionDeniedError represents a 403 error
type PermissionDeniedError struct {
	APIError
}

// NotFoundError represents a 404 error
type NotFoundError struct {
	APIError
}

// ConflictError represents a 409 error
type ConflictError struct {
	APIError
}

// UnprocessableEntityError represents a 422 error
type UnprocessableEntityError struct {
	APIError
}

// RateLimitError represents a 429 error
type RateLimitError struct {
	APIError
}

// InternalServerError represents a 5xx error
type InternalServerError struct {
	APIError
}

// IsRetryableError returns true if the error is retryable
func IsRetryableError(err error) bool {
	switch err.(type) {
	case *APIConnectionError, *APIConnectionTimeoutError:
		return true
	case *ConflictError, *RateLimitError, *InternalServerError:
		return true
	case *APIError:
		apiErr := err.(*APIError)
		return apiErr.Status == 408 || apiErr.Status >= 500
	default:
		return false
	}
}
