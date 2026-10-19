package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// maxResponseBytes caps how much of a response body is read into memory.
const maxResponseBytes = 16 << 20

// APIError is returned when a vendor API responds with a non-2xx status.
type APIError struct {
	Provider   string
	StatusCode int
	Type       string // vendor error type/status, when provided
	Message    string
}

func (e *APIError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("llm: %s api error (status %d, %s): %s", e.Provider, e.StatusCode, e.Type, e.Message)
	}
	return fmt.Sprintf("llm: %s api error (status %d): %s", e.Provider, e.StatusCode, e.Message)
}

// Retryable reports whether the request may succeed if sent again (rate limit or server-side failure).
func (e *APIError) Retryable() bool {
	return e.StatusCode == http.StatusTooManyRequests ||
		e.StatusCode == http.StatusRequestTimeout ||
		e.StatusCode >= http.StatusInternalServerError
}

// errorDecoder turns a non-2xx response body into a vendor-specific APIError.
type errorDecoder func(statusCode int, body []byte) *APIError

// postJSON sends payload as JSON and decodes a successful response into out.
func postJSON(ctx context.Context, client *http.Client, url string, headers map[string]string, payload, out any, decodeErr errorDecoder) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("llm: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("llm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("llm: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("llm: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeErr(resp.StatusCode, respBody)
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("llm: decode response: %w", err)
	}
	return nil
}

// fallbackMessage returns a short message for error bodies that could not be parsed.
func fallbackMessage(statusCode int, body []byte) string {
	const limit = 512
	if len(body) == 0 {
		return http.StatusText(statusCode)
	}
	if len(body) > limit {
		return string(body[:limit]) + "..."
	}
	return string(body)
}
