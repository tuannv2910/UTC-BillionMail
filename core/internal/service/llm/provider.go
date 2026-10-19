// Package llm provides a provider-agnostic abstraction over large language model APIs
// (OpenAI, Anthropic Claude, Google Gemini, and OpenAI-compatible endpoints).
//
// Callers depend only on the Provider interface; concrete implementations are created
// through a Registry so new vendors can be added without touching existing code.
package llm

import (
	"context"
	"errors"
)

// Provider is the single abstraction business code depends on to generate content.
type Provider interface {
	// Name returns the registered provider name (e.g. "openai", "anthropic", "gemini").
	Name() string
	// Generate sends the conversation to the model and returns the complete response.
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
}

// Role identifies the author of a conversation message.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is one turn of the conversation. System instructions go in GenerateRequest.System.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// GenerateRequest is the provider-neutral input for a generation call.
type GenerateRequest struct {
	Model       string    // overrides Config.Model when set
	System      string    // system instruction, mapped to each vendor's native field
	Messages    []Message // conversation turns, must end with a user message
	MaxTokens   int       // overrides Config.MaxTokens when > 0
	Temperature *float64  // nil leaves the vendor default
}

// FinishReason is the normalized reason the model stopped generating.
type FinishReason string

const (
	FinishStop          FinishReason = "stop"
	FinishLength        FinishReason = "length"
	FinishContentFilter FinishReason = "content_filter"
	FinishOther         FinishReason = "other"
)

// Usage reports token consumption for a single call.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// GenerateResponse is the provider-neutral result of a generation call.
type GenerateResponse struct {
	Content      string       `json:"content"`
	Model        string       `json:"model"`
	FinishReason FinishReason `json:"finish_reason"`
	Usage        Usage        `json:"usage"`
}

var (
	ErrUnknownProvider = errors.New("llm: unknown provider")
	ErrInvalidConfig   = errors.New("llm: invalid config")
	ErrInvalidRequest  = errors.New("llm: invalid request")
	ErrEmptyResponse   = errors.New("llm: empty response")
)

// Float64 returns a pointer to v, for setting GenerateRequest.Temperature inline.
func Float64(v float64) *float64 {
	return &v
}

// NewPrompt builds a single-turn request from a system instruction and a user prompt.
func NewPrompt(system, prompt string) *GenerateRequest {
	return &GenerateRequest{
		System:   system,
		Messages: []Message{{Role: RoleUser, Content: prompt}},
	}
}

// validate checks the request shape shared by every provider.
func (r *GenerateRequest) validate() error {
	if r == nil || len(r.Messages) == 0 {
		return errors.Join(ErrInvalidRequest, errors.New("at least one message is required"))
	}
	for _, m := range r.Messages {
		if m.Role != RoleUser && m.Role != RoleAssistant {
			return errors.Join(ErrInvalidRequest, errors.New("unsupported message role: "+string(m.Role)))
		}
	}
	if r.Messages[len(r.Messages)-1].Role != RoleUser {
		return errors.Join(ErrInvalidRequest, errors.New("last message must be from the user"))
	}
	return nil
}
