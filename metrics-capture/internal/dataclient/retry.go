package dataclient

import (
	"context"
	"time"
)

// RetryConfig controls lightweight retry behavior for uploads.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
}

func (c RetryConfig) normalized() RetryConfig {
	out := c
	if out.MaxAttempts <= 0 {
		out.MaxAttempts = 1
	}
	if out.BaseDelay <= 0 {
		out.BaseDelay = 250 * time.Millisecond
	}
	return out
}

func retryUpload(ctx context.Context, cfg RetryConfig, fn func(context.Context) (string, error)) (string, error) {
	cfg = cfg.normalized()

	var lastErr error
	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		fileID, err := fn(ctx)
		if err == nil {
			return fileID, nil
		}
		lastErr = err

		if attempt == cfg.MaxAttempts {
			break
		}
		delay := cfg.BaseDelay * time.Duration(1<<(attempt-1))
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return "", ctx.Err()
		case <-timer.C:
		}
	}
	return "", lastErr
}
