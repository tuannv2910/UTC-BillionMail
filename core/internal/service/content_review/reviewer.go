package content_review

import (
	"billionmail-core/internal/service/llm"
	"context"
	"errors"
	"strings"
)

const defaultMaxTokens = 4096

// Reviewer reviews email content with any llm.Provider.
type Reviewer struct {
	provider  llm.Provider
	model     string
	maxTokens int
}

// Option customizes a Reviewer.
type Option func(*Reviewer)

// WithModel overrides the provider's default model.
func WithModel(model string) Option { return func(r *Reviewer) { r.model = model } }

// WithMaxTokens caps the reply length.
func WithMaxTokens(n int) Option { return func(r *Reviewer) { r.maxTokens = n } }

// NewReviewer returns a Reviewer backed by provider.
func NewReviewer(provider llm.Provider, opts ...Option) *Reviewer {
	r := &Reviewer{provider: provider, maxTokens: defaultMaxTokens}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Review asks the model for issues, keeps those that quote real text from the email, and scores them.
// A reply that is not valid JSON is retried once with a repair instruction.
func (r *Reviewer) Review(ctx context.Context, in Input) (*Result, error) {
	if strings.TrimSpace(in.Subject) == "" && strings.TrimSpace(in.Content) == "" {
		return nil, ErrEmptyInput
	}

	prepared := prepareEmail(in.Subject, in.Content)
	req := &llm.GenerateRequest{
		Model:       r.model,
		System:      buildSystemPrompt(in),
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: buildUserPrompt(prepared)}},
		MaxTokens:   r.maxTokens,
		Temperature: llm.Float64(0),
	}

	var usage llm.Usage
	var parsed *modelReview
	var model string
	for attempt := 0; attempt < 2; attempt++ {
		resp, err := r.provider.Generate(ctx, req)
		if err != nil {
			return nil, err
		}
		usage.InputTokens += resp.Usage.InputTokens
		usage.OutputTokens += resp.Usage.OutputTokens
		model = resp.Model

		parsed, err = parseModelReview(resp.Content)
		if err == nil {
			break
		}
		if !errors.Is(err, ErrInvalidModelOutput) || attempt == 1 {
			return nil, err
		}
		req.Messages = append(req.Messages,
			llm.Message{Role: llm.RoleAssistant, Content: resp.Content},
			llm.Message{Role: llm.RoleUser, Content: repairPrompt},
		)
	}

	result := buildResult(parsed, prepared, in.Content)
	result.Model = model
	result.PromptVersion = PromptVersion
	result.InputTokens = usage.InputTokens
	result.OutputTokens = usage.OutputTokens
	return result, nil
}
