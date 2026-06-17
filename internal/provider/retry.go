// Package provider implements retry logic with exponential backoff.
package provider

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// RetryConfig controls retry behavior.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig returns sensible defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 5,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   5 * time.Minute,
	}
}

// IsRetryable checks if an error is transient and safe to retry.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()

	// HTTP status codes
	for _, code := range []int{429, 500, 502, 503, 504, 529} {
		if strings.Contains(msg, fmt.Sprintf("API error %d", code)) {
			return true
		}
	}

	// Network errors
	retryable := []string{"ECONNRESET", "EPIPE", "ETIMEDOUT", "ENETUNREACH", "EAI_AGAIN", "connection refused"}
	for _, r := range retryable {
		if strings.Contains(msg, r) {
			return true
		}
	}

	// SSE timeout
	if strings.Contains(msg, "SSE read timed out") || strings.Contains(msg, "ECONNABORTED") {
		return true
	}

	return false
}

// IsRateLimited checks if the error is a rate limit (429).
func IsRateLimited(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "API error 429")
}

// ExtractRetryAfter extracts Retry-After from an HTTP response header.
func ExtractRetryAfter(header http.Header) time.Duration {
	if ra := header.Get("Retry-After"); ra != "" {
		var seconds int
		if _, err := fmt.Sscanf(ra, "%d", &seconds); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return 0
}

// CalculateDelay computes the delay for a given retry attempt with jitter.
func CalculateDelay(attempt int, cfg RetryConfig, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		if retryAfter > cfg.MaxDelay {
			return cfg.MaxDelay
		}
		return retryAfter
	}

	exp := float64(cfg.BaseDelay) * math.Pow(2, float64(attempt))
	jitter := exp * 0.25 * (rand.Float64()*2 - 1)
	delay := time.Duration(exp + jitter)
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}
	return delay
}

// WithRetry executes a function with retry logic.
func WithRetry(cfg RetryConfig, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		if !IsRetryable(err) || attempt >= cfg.MaxRetries {
			return err
		}

		delay := CalculateDelay(attempt, cfg, 0)
		time.Sleep(delay)
	}
	return lastErr
}

// WithRetryResult is a generic retry function that returns a result.
func WithRetryResult[T any](cfg RetryConfig, fn func() (T, error)) (T, error) {
	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}
		lastErr = err

		if !IsRetryable(err) || attempt >= cfg.MaxRetries {
			var zero T
			return zero, err
		}

		delay := CalculateDelay(attempt, cfg, 0)
		time.Sleep(delay)
	}
	var zero T
	return zero, lastErr
}
