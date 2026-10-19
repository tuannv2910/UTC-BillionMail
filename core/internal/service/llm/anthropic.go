package llm

import (
	"context"
	"encoding/json"
	"strings"
)

const (
	anthropicDefaultBaseURL = "https://api.anthropic.com/v1"
	anthropicAPIVersion     = "2023-06-01"
)

// anthropicProvider talks to the native Claude Messages API.
type anthropicProvider struct {
	cfg Config
}

// NewAnthropic builds a provider for Anthropic Claude models.
func NewAnthropic(cfg Config) (Provider, error) {
	cfg = cfg.withDefaults(anthropicDefaultBaseURL)
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &anthropicProvider{cfg: cfg}, nil
}

func (p *anthropicProvider) Name() string { return ProviderAnthropic }

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	Temperature *float64           `json:"temperature,omitempty"`
}

type anthropicResponse struct {
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (p *anthropicProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	model, maxTokens := p.cfg.resolve(req)
	body := anthropicRequest{
		Model:       model,
		MaxTokens:   maxTokens,
		System:      req.System,
		Messages:    make([]anthropicMessage, 0, len(req.Messages)),
		Temperature: req.Temperature,
	}
	for _, m := range req.Messages {
		body.Messages = append(body.Messages, anthropicMessage{Role: string(m.Role), Content: m.Content})
	}

	headers := map[string]string{
		"x-api-key":         p.cfg.APIKey,
		"anthropic-version": anthropicAPIVersion,
	}
	var resp anthropicResponse
	if err := postJSON(ctx, p.cfg.HTTPClient, p.cfg.BaseURL+"/messages", headers, body, &resp, decodeAnthropicError); err != nil {
		return nil, err
	}

	// Only text blocks carry the answer; thinking and tool blocks are skipped.
	var sb strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	if sb.Len() == 0 {
		return nil, ErrEmptyResponse
	}

	return &GenerateResponse{
		Content:      sb.String(),
		Model:        resp.Model,
		FinishReason: anthropicFinishReason(resp.StopReason),
		Usage: Usage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
			TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}, nil
}

func decodeAnthropicError(statusCode int, body []byte) *APIError {
	var payload struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	apiErr := &APIError{Provider: ProviderAnthropic, StatusCode: statusCode}
	if json.Unmarshal(body, &payload) == nil && payload.Error.Message != "" {
		apiErr.Type = payload.Error.Type
		apiErr.Message = payload.Error.Message
		return apiErr
	}
	apiErr.Message = fallbackMessage(statusCode, body)
	return apiErr
}

func anthropicFinishReason(reason string) FinishReason {
	switch reason {
	case "end_turn", "stop_sequence":
		return FinishStop
	case "max_tokens":
		return FinishLength
	case "refusal":
		return FinishContentFilter
	default:
		return FinishOther
	}
}
