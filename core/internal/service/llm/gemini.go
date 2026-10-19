package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const geminiDefaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// geminiProvider talks to the native Gemini generateContent API.
type geminiProvider struct {
	cfg Config
}

// NewGemini builds a provider for Google Gemini models.
func NewGemini(cfg Config) (Provider, error) {
	cfg = cfg.withDefaults(geminiDefaultBaseURL)
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &geminiProvider{cfg: cfg}, nil
}

func (p *geminiProvider) Name() string { return ProviderGemini }

type geminiPart struct {
	Text    string `json:"text,omitempty"`
	Thought bool   `json:"thought,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	Temperature     *float64 `json:"temperature,omitempty"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent         `json:"systemInstruction,omitempty"`
	Contents          []geminiContent        `json:"contents"`
	GenerationConfig  geminiGenerationConfig `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content      geminiContent `json:"content"`
		FinishReason string        `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
	ModelVersion string `json:"modelVersion"`
}

func (p *geminiProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	model, maxTokens := p.cfg.resolve(req)
	body := geminiRequest{
		Contents: make([]geminiContent, 0, len(req.Messages)),
		GenerationConfig: geminiGenerationConfig{
			MaxOutputTokens: maxTokens,
			Temperature:     req.Temperature,
		},
	}
	if req.System != "" {
		body.SystemInstruction = &geminiContent{Parts: []geminiPart{{Text: req.System}}}
	}
	for _, m := range req.Messages {
		role := "user"
		if m.Role == RoleAssistant {
			role = "model"
		}
		body.Contents = append(body.Contents, geminiContent{Role: role, Parts: []geminiPart{{Text: m.Content}}})
	}

	endpoint := fmt.Sprintf("%s/models/%s:generateContent", p.cfg.BaseURL, url.PathEscape(strings.TrimPrefix(model, "models/")))
	headers := map[string]string{"x-goog-api-key": p.cfg.APIKey}
	var resp geminiResponse
	if err := postJSON(ctx, p.cfg.HTTPClient, endpoint, headers, body, &resp, decodeGeminiError); err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 {
		if resp.PromptFeedback.BlockReason != "" {
			return nil, fmt.Errorf("%w: prompt blocked (%s)", ErrEmptyResponse, resp.PromptFeedback.BlockReason)
		}
		return nil, ErrEmptyResponse
	}

	// Thought parts are internal reasoning from thinking models, not the answer.
	candidate := resp.Candidates[0]
	var sb strings.Builder
	for _, part := range candidate.Content.Parts {
		if !part.Thought {
			sb.WriteString(part.Text)
		}
	}
	if sb.Len() == 0 {
		return nil, fmt.Errorf("%w: finish reason %s", ErrEmptyResponse, candidate.FinishReason)
	}

	modelName := resp.ModelVersion
	if modelName == "" {
		modelName = model
	}
	return &GenerateResponse{
		Content:      sb.String(),
		Model:        modelName,
		FinishReason: geminiFinishReason(candidate.FinishReason),
		Usage: Usage{
			InputTokens:  resp.UsageMetadata.PromptTokenCount,
			OutputTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:  resp.UsageMetadata.TotalTokenCount,
		},
	}, nil
}

func decodeGeminiError(statusCode int, body []byte) *APIError {
	var payload struct {
		Error struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	apiErr := &APIError{Provider: ProviderGemini, StatusCode: statusCode}
	if json.Unmarshal(body, &payload) == nil && payload.Error.Message != "" {
		apiErr.Type = payload.Error.Status
		apiErr.Message = payload.Error.Message
		return apiErr
	}
	apiErr.Message = fallbackMessage(statusCode, body)
	return apiErr
}

func geminiFinishReason(reason string) FinishReason {
	switch reason {
	case "STOP":
		return FinishStop
	case "MAX_TOKENS":
		return FinishLength
	case "SAFETY", "RECITATION", "BLOCKLIST", "PROHIBITED_CONTENT", "SPII":
		return FinishContentFilter
	default:
		return FinishOther
	}
}
