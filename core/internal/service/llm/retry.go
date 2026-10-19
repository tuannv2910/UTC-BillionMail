package llm

import (
	"context"
	"errors"
	"time"
)

// retryProvider decorates a Provider, re-sending requests that fail with a retryable APIError.
type retryProvider struct {
	next        Provider
	maxAttempts int
	baseDelay   time.Duration
}

// WithRetry wraps p so rate-limit and 5xx errors are retried up to maxAttempts total,
// with exponential backoff starting at baseDelay. Other errors are returned immediately.
func WithRetry(p Provider, maxAttempts int, baseDelay time.Duration) Provider {
	if maxAttempts <= 1 {
		return p
	}
	return &retryProvider{next: p, maxAttempts: maxAttempts, baseDelay: baseDelay}
}

func (r *retryProvider) Name() string { return r.next.Name() }

func (r *retryProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	delay := r.baseDelay
	for attempt := 1; ; attempt++ {
		resp, err := r.next.Generate(ctx, req)
		if err == nil || attempt >= r.maxAttempts || !isRetryable(err) {
			return resp, err
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
		delay *= 2
	}
}

func isRetryable(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Retryable()
}
