package llm

import (
	"context"
	"encoding/json"
)

const openAIDefaultBaseURL = "https://api.openai.com/v1"

// openAIProvider talks to the Chat Completions API. It also serves OpenAI-compatible
// vendors (DeepSeek, Grok, Kimi, ...) when built with a custom BaseURL.
type openAIProvider struct {
	name string
	cfg  Config
	// legacyMaxTokens sends "max_tokens" instead of "max_completion_tokens";
	// most OpenAI-compatible vendors only understand the former.
	legacyMaxTokens bool
}

// NewOpenAI builds a provider for the official OpenAI API.
func NewOpenAI(cfg Config) (Provider, error) {
	return newOpenAIProvider(ProviderOpenAI, cfg, openAIDefaultBaseURL, false)
}

// NewOpenAICompatible builds a provider for any vendor exposing an OpenAI-style
// /chat/completions endpoint. BaseURL is required.
func NewOpenAICompatible(cfg Config) (Provider, error) {
	return newOpenAIProvider(ProviderOpenAICompatible, cfg, "", true)
}

func newOpenAIProvider(name string, cfg Config, defaultBaseURL string, legacyMaxTokens bool) (Provider, error) {
	cfg = cfg.withDefaults(defaultBaseURL)
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &openAIProvider{name: name, cfg: cfg, legacyMaxTokens: legacyMaxTokens}, nil
}

func (p *openAIProvider) Name() string { return p.name }

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model               string          `json:"model"`
	Messages            []openAIMessage `json:"messages"`
	MaxCompletionTokens int             `json:"max_completion_tokens,omitempty"`
	MaxTokens           int             `json:"max_tokens,omitempty"`
	Temperature         *float64        `json:"temperature,omitempty"`
}

type openAIResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (p *openAIProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	model, maxTokens := p.cfg.resolve(req)
	body := openAIRequest{
		Model:       model,
		Messages:    make([]openAIMessage, 0, len(req.Messages)+1),
		Temperature: req.Temperature,
	}
	if p.legacyMaxTokens {
		body.MaxTokens = maxTokens
	} else {
		body.MaxCompletionTokens = maxTokens
	}
	if req.System != "" {
		body.Messages = append(body.Messages, openAIMessage{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		body.Messages = append(body.Messages, openAIMessage{Role: string(m.Role), Content: m.Content})
	}

	headers := map[string]string{"Authorization": "Bearer " + p.cfg.APIKey}
	var resp openAIResponse
	if err := postJSON(ctx, p.cfg.HTTPClient, p.cfg.BaseURL+"/chat/completions", headers, body, &resp, p.decodeError); err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return nil, ErrEmptyResponse
	}
	choice := resp.Choices[0]
	return &GenerateResponse{
		Content:      choice.Message.Content,
		Model:        resp.Model,
		FinishReason: openAIFinishReason(choice.FinishReason),
		Usage: Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
	}, nil
}

func (p *openAIProvider) decodeError(statusCode int, body []byte) *APIError {
	var payload struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	apiErr := &APIError{Provider: p.name, StatusCode: statusCode}
	if json.Unmarshal(body, &payload) == nil && payload.Error.Message != "" {
		apiErr.Type = payload.Error.Type
		apiErr.Message = payload.Error.Message
		return apiErr
	}
	apiErr.Message = fallbackMessage(statusCode, body)
	return apiErr
}

func openAIFinishReason(reason string) FinishReason {
	switch reason {
	case "stop":
		return FinishStop
	case "length":
		return FinishLength
	case "content_filter":
		return FinishContentFilter
	default:
		return FinishOther
	}
}
