// Package retry provides a retry wrapper for functions that may fail intermittently.
// It supports configurable retry attempts, delays, and exponential backoff.
package retry

import (
	"easyflow-oauth2-server/pkg/logger"
	"time"
)

// Config holds the configuration for retry behavior.
type Config struct {
	// MaxAttempts is the maximum number of attempts before giving up
	MaxAttempts int
	// Delay is the initial delay between attempts
	Delay time.Duration
	// MaxDelay is the maximum delay between attempts
	MaxDelay time.Duration
	// Multiplier is the factor by which the delay is multiplied between attempts
	Multiplier float64
	// FunctionName is the name of the function being retried, used for logging
	FunctionName string
	// RetryableErr is a function that determines if an error is retryable
	RetryableErr func(error) bool
}

// DefaultRetryConfig provides sensible default values.
func DefaultRetryConfig(name string) Config {
	return Config{
		MaxAttempts:  5,
		Delay:        time.Second,
		MaxDelay:     time.Second * 30,
		Multiplier:   2.0,
		FunctionName: name,
		RetryableErr: func(_ error) bool {
			return true // By default, retry all errors
		},
	}
}

// WithRetry wraps a function with retry logic.
func WithRetry[T any](
	fn func() (T, error),
	logger *logger.Logger,
	config Config,
) func() (T, error) {
	return func() (T, error) {
		var lastErr error
		currentDelay := config.Delay

		for attempt := range config.MaxAttempts {
			result, err := fn()
			if err == nil {
				return result, nil
			}

			lastErr = err
			if !config.RetryableErr(err) {
				var zero T
				return zero, err
			}

			if attempt < config.MaxAttempts-1 {
				time.Sleep(currentDelay)
				currentDelay = min(
					time.Duration(float64(currentDelay)*config.Multiplier),
					config.MaxDelay,
				)
				logger.PrintfWarning(
					"Failed to complete function %s successfully retring again in %f. Attempt %d",
					config.FunctionName,
					currentDelay.Seconds(),
					attempt,
				)
			}
		}

		var zero T
		logger.PrintfError("Reached max retry attempts for func: %s", config.FunctionName)
		return zero, lastErr
	}
}
