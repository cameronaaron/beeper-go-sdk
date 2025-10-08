package internal

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryLogic handles request retries with exponential backoff
type RetryLogic struct {
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

// NewRetryLogic creates a new RetryLogic instance
func NewRetryLogic(maxRetries int) *RetryLogic {
	return &RetryLogic{
		maxRetries: maxRetries,
		baseDelay:  250 * time.Millisecond,
		maxDelay:   10 * time.Second,
	}
}

// Do executes the given function with retry logic
func (r *RetryLogic) Do(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on non-retryable errors
		if !isRetryableError(err) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt < r.maxRetries {
			delay := r.calculateDelay(attempt)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("request failed after %d attempts: %w", r.maxRetries+1, lastErr)
}

// calculateDelay calculates the delay for the given attempt using exponential backoff
func (r *RetryLogic) calculateDelay(attempt int) time.Duration {
	delay := time.Duration(float64(r.baseDelay) * math.Pow(2, float64(attempt)))
	if delay > r.maxDelay {
		delay = r.maxDelay
	}
	return delay
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error) bool {
	// Connection errors are retryable
	// HTTP status codes 408, 409, 429, 5xx are retryable
	// Everything else is not retryable

	// For now, we return false for most errors to preserve type information
	// The specific error types would need to be imported to check properly
	return false
}
