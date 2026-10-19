package llm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// capturedRequest records what a fake vendor server received.
type capturedRequest struct {
	Path    string
	Headers http.Header
	Body    map[string]any
}

// fakeServer replies with status/body and records the incoming request.
func fakeServer(t *testing.T, status int, body string) (*httptest.Server, *capturedRequest) {
	t.Helper()
	captured := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.Path = r.URL.Path
		captured.Headers = r.Header.Clone()
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &captured.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, captured
}

func testConfig(provider, baseURL string) Config {
	return Config{Provider: provider, APIKey: "test-key", BaseURL: baseURL, Model: "test-model"}
}

func TestOpenAI_Generate(t *testing.T) {
	srv, got := fakeServer(t, http.StatusOK, `{
		"model": "gpt-test",
		"choices": [{"message": {"content": "<h1>Hi</h1>"}, "finish_reason": "stop"}],
		"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
	}`)

	p, err := New(testConfig(ProviderOpenAI, srv.URL))
	require.NoError(t, err)

	req := NewPrompt("You write emails.", "Write a welcome email.")
	req.Temperature = Float64(0.3)
	resp, err := p.Generate(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, "/chat/completions", got.Path)
	assert.Equal(t, "Bearer test-key", got.Headers.Get("Authorization"))
	assert.Equal(t, "test-model", got.Body["model"])
	assert.EqualValues(t, defaultMaxTokens, got.Body["max_completion_tokens"])
	assert.NotContains(t, got.Body, "max_tokens")
	assert.EqualValues(t, 0.3, got.Body["temperature"])
	messages := got.Body["messages"].([]any)
	require.Len(t, messages, 2)
	assert.Equal(t, "system", messages[0].(map[string]any)["role"])
	assert.Equal(t, "user", messages[1].(map[string]any)["role"])

	assert.Equal(t, "<h1>Hi</h1>", resp.Content)
	assert.Equal(t, "gpt-test", resp.Model)
	assert.Equal(t, FinishStop, resp.FinishReason)
	assert.Equal(t, Usage{InputTokens: 10, OutputTokens: 5, TotalTokens: 15}, resp.Usage)
}

func TestOpenAICompatible_UsesLegacyMaxTokens(t *testing.T) {
	srv, got := fakeServer(t, http.StatusOK, `{"choices": [{"message": {"content": "ok"}, "finish_reason": "length"}]}`)

	p, err := New(testConfig(ProviderOpenAICompatible, srv.URL+"/"))
	require.NoError(t, err)
	assert.Equal(t, ProviderOpenAICompatible, p.Name())

	resp, err := p.Generate(context.Background(), &GenerateRequest{
		Model:     "deepseek-chat",
		MaxTokens: 256,
		Messages:  []Message{{Role: RoleUser, Content: "hi"}},
	})
	require.NoError(t, err)

	assert.Equal(t, "/chat/completions", got.Path)
	assert.Equal(t, "deepseek-chat", got.Body["model"])
	assert.EqualValues(t, 256, got.Body["max_tokens"])
	assert.NotContains(t, got.Body, "max_completion_tokens")
	assert.NotContains(t, got.Body, "temperature")
	assert.Equal(t, FinishLength, resp.FinishReason)
}

func TestOpenAICompatible_RequiresBaseURL(t *testing.T) {
	_, err := New(Config{Provider: ProviderOpenAICompatible, APIKey: "k", Model: "m"})
	assert.ErrorIs(t, err, ErrInvalidConfig)
}

func TestAnthropic_Generate(t *testing.T) {
	srv, got := fakeServer(t, http.StatusOK, `{
		"model": "claude-test",
		"content": [
			{"type": "thinking", "thinking": "internal"},
			{"type": "text", "text": "Hello "},
			{"type": "text", "text": "there"}
		],
		"stop_reason": "end_turn",
		"usage": {"input_tokens": 12, "output_tokens": 4}
	}`)

	p, err := New(testConfig(ProviderAnthropic, srv.URL))
	require.NoError(t, err)

	resp, err := p.Generate(context.Background(), &GenerateRequest{
		System: "You write emails.",
		Messages: []Message{
			{Role: RoleUser, Content: "Draft a subject line"},
			{Role: RoleAssistant, Content: "Big news!"},
			{Role: RoleUser, Content: "Shorter"},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "/messages", got.Path)
	assert.Equal(t, "test-key", got.Headers.Get("x-api-key"))
	assert.Equal(t, anthropicAPIVersion, got.Headers.Get("anthropic-version"))
	assert.Equal(t, "You write emails.", got.Body["system"])
	assert.EqualValues(t, defaultMaxTokens, got.Body["max_tokens"])
	assert.Len(t, got.Body["messages"], 3)

	assert.Equal(t, "Hello there", resp.Content)
	assert.Equal(t, "claude-test", resp.Model)
	assert.Equal(t, FinishStop, resp.FinishReason)
	assert.Equal(t, Usage{InputTokens: 12, OutputTokens: 4, TotalTokens: 16}, resp.Usage)
}

func TestAnthropic_APIError(t *testing.T) {
	srv, _ := fakeServer(t, http.StatusTooManyRequests,
		`{"type": "error", "error": {"type": "rate_limit_error", "message": "slow down"}}`)

	p, err := New(testConfig(ProviderAnthropic, srv.URL))
	require.NoError(t, err)

	_, err = p.Generate(context.Background(), NewPrompt("", "hi"))
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, ProviderAnthropic, apiErr.Provider)
	assert.Equal(t, http.StatusTooManyRequests, apiErr.StatusCode)
	assert.Equal(t, "rate_limit_error", apiErr.Type)
	assert.Equal(t, "slow down", apiErr.Message)
	assert.True(t, apiErr.Retryable())
}

func TestGemini_Generate(t *testing.T) {
	srv, got := fakeServer(t, http.StatusOK, `{
		"candidates": [{
			"content": {"role": "model", "parts": [
				{"text": "plan...", "thought": true},
				{"text": "<p>Welcome</p>"}
			]},
			"finishReason": "STOP"
		}],
		"usageMetadata": {"promptTokenCount": 7, "candidatesTokenCount": 3, "totalTokenCount": 10},
		"modelVersion": "gemini-test-001"
	}`)

	p, err := New(testConfig(ProviderGemini, srv.URL))
	require.NoError(t, err)

	resp, err := p.Generate(context.Background(), &GenerateRequest{
		Model:  "models/gemini-test",
		System: "You write emails.",
		Messages: []Message{
			{Role: RoleUser, Content: "Draft"},
			{Role: RoleAssistant, Content: "Draft v1"},
			{Role: RoleUser, Content: "Improve"},
		},
		Temperature: Float64(0.5),
	})
	require.NoError(t, err)

	assert.Equal(t, "/models/gemini-test:generateContent", got.Path)
	assert.Equal(t, "test-key", got.Headers.Get("x-goog-api-key"))
	system := got.Body["systemInstruction"].(map[string]any)
	assert.Equal(t, "You write emails.", system["parts"].([]any)[0].(map[string]any)["text"])
	contents := got.Body["contents"].([]any)
	require.Len(t, contents, 3)
	assert.Equal(t, "model", contents[1].(map[string]any)["role"])
	genCfg := got.Body["generationConfig"].(map[string]any)
	assert.EqualValues(t, defaultMaxTokens, genCfg["maxOutputTokens"])
	assert.EqualValues(t, 0.5, genCfg["temperature"])

	assert.Equal(t, "<p>Welcome</p>", resp.Content)
	assert.Equal(t, "gemini-test-001", resp.Model)
	assert.Equal(t, FinishStop, resp.FinishReason)
	assert.Equal(t, Usage{InputTokens: 7, OutputTokens: 3, TotalTokens: 10}, resp.Usage)
}

func TestGemini_PromptBlocked(t *testing.T) {
	srv, _ := fakeServer(t, http.StatusOK, `{"promptFeedback": {"blockReason": "SAFETY"}}`)

	p, err := New(testConfig(ProviderGemini, srv.URL))
	require.NoError(t, err)

	_, err = p.Generate(context.Background(), NewPrompt("", "hi"))
	assert.ErrorIs(t, err, ErrEmptyResponse)
	assert.Contains(t, err.Error(), "SAFETY")
}

func TestGemini_APIError(t *testing.T) {
	srv, _ := fakeServer(t, http.StatusBadRequest,
		`{"error": {"code": 400, "message": "API key not valid", "status": "INVALID_ARGUMENT"}}`)

	p, err := New(testConfig(ProviderGemini, srv.URL))
	require.NoError(t, err)

	_, err = p.Generate(context.Background(), NewPrompt("", "hi"))
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "INVALID_ARGUMENT", apiErr.Type)
	assert.False(t, apiErr.Retryable())
}

func TestAPIError_UnparseableBody(t *testing.T) {
	srv, _ := fakeServer(t, http.StatusBadGateway, `<html>bad gateway</html>`)

	p, err := New(testConfig(ProviderOpenAI, srv.URL))
	require.NoError(t, err)

	_, err = p.Generate(context.Background(), NewPrompt("", "hi"))
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "<html>bad gateway</html>", apiErr.Message)
	assert.True(t, apiErr.Retryable())
}

func TestEmptyResponse(t *testing.T) {
	srv, _ := fakeServer(t, http.StatusOK, `{"choices": []}`)

	p, err := New(testConfig(ProviderOpenAI, srv.URL))
	require.NoError(t, err)

	_, err = p.Generate(context.Background(), NewPrompt("", "hi"))
	assert.ErrorIs(t, err, ErrEmptyResponse)
}

func TestConfigValidation(t *testing.T) {
	cases := map[string]Config{
		"missing api key": {Provider: ProviderOpenAI, Model: "m"},
		"missing model":   {Provider: ProviderAnthropic, APIKey: "k"},
	}
	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := New(cfg)
			assert.ErrorIs(t, err, ErrInvalidConfig)
		})
	}
}

func TestRequestValidation(t *testing.T) {
	p, err := New(Config{Provider: ProviderGemini, APIKey: "k", Model: "m"})
	require.NoError(t, err)

	cases := map[string]*GenerateRequest{
		"nil request":       nil,
		"no messages":       {System: "s"},
		"bad role":          {Messages: []Message{{Role: "system", Content: "x"}}},
		"ends on assistant": {Messages: []Message{{Role: RoleUser, Content: "a"}, {Role: RoleAssistant, Content: "b"}}},
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := p.Generate(context.Background(), req)
			assert.ErrorIs(t, err, ErrInvalidRequest)
		})
	}
}

func TestRegistry(t *testing.T) {
	assert.Equal(t, []string{ProviderAnthropic, ProviderGemini, ProviderOpenAI, ProviderOpenAICompatible}, Names())

	_, err := New(Config{Provider: "unknown"})
	assert.ErrorIs(t, err, ErrUnknownProvider)

	// Names are case-insensitive so values from config files or the UI work as-is.
	p, err := New(Config{Provider: " Gemini ", APIKey: "k", Model: "m"})
	require.NoError(t, err)
	assert.Equal(t, ProviderGemini, p.Name())

	// A custom registry accepts new vendors without changing existing code.
	r := NewRegistry()
	r.Register("stub", func(Config) (Provider, error) { return &stubProvider{}, nil })
	p, err = r.New(Config{Provider: "STUB"})
	require.NoError(t, err)
	assert.Equal(t, "stub", p.Name())
}

// stubProvider fails with the queued errors in order, then succeeds.
type stubProvider struct {
	errs  []error
	calls atomic.Int32
}

func (s *stubProvider) Name() string { return "stub" }

func (s *stubProvider) Generate(context.Context, *GenerateRequest) (*GenerateResponse, error) {
	n := int(s.calls.Add(1))
	if n <= len(s.errs) {
		return nil, s.errs[n-1]
	}
	return &GenerateResponse{Content: "ok"}, nil
}

func TestWithRetry_RetriesRetryableErrors(t *testing.T) {
	stub := &stubProvider{errs: []error{
		&APIError{StatusCode: http.StatusTooManyRequests},
		&APIError{StatusCode: http.StatusServiceUnavailable},
	}}
	p := WithRetry(stub, 3, time.Millisecond)

	resp, err := p.Generate(context.Background(), NewPrompt("", "hi"))
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Content)
	assert.EqualValues(t, 3, stub.calls.Load())
	assert.Equal(t, "stub", p.Name())
}

func TestWithRetry_StopsOnNonRetryable(t *testing.T) {
	stub := &stubProvider{errs: []error{&APIError{StatusCode: http.StatusUnauthorized}}}
	p := WithRetry(stub, 3, time.Millisecond)

	_, err := p.Generate(context.Background(), NewPrompt("", "hi"))
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
	assert.EqualValues(t, 1, stub.calls.Load())
}

func TestWithRetry_GivesUpAfterMaxAttempts(t *testing.T) {
	retryable := &APIError{StatusCode: http.StatusInternalServerError}
	stub := &stubProvider{errs: []error{retryable, retryable, retryable}}
	p := WithRetry(stub, 2, time.Millisecond)

	_, err := p.Generate(context.Background(), NewPrompt("", "hi"))
	assert.ErrorIs(t, err, retryable)
	assert.EqualValues(t, 2, stub.calls.Load())
}

func TestWithRetry_HonorsContextCancel(t *testing.T) {
	stub := &stubProvider{errs: []error{&APIError{StatusCode: http.StatusTooManyRequests}}}
	p := WithRetry(stub, 3, time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := p.Generate(ctx, NewPrompt("", "hi"))
	assert.True(t, errors.Is(err, context.Canceled))
	assert.EqualValues(t, 1, stub.calls.Load())
}
