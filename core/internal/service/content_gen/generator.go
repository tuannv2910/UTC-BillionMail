// Package content_gen writes or revises HTML marketing emails with any llm.Provider.
package content_gen

import (
	"billionmail-core/internal/service/llm"
	"context"
	"errors"
	"strings"
)

const (
	defaultMaxTokens = 8192
	maxBriefRunes    = 4000
	maxCurrentRunes  = 60000
)

var (
	ErrEmptyBrief         = errors.New("content_gen: brief or instruction is required")
	ErrInvalidModelOutput = errors.New("content_gen: model did not return an HTML email")
)

// Input describes the email to write. When CurrentHTML is set the generator revises that email
// following Brief instead of writing a new one.
type Input struct {
	Brief       string // what the email is about, or how to change CurrentHTML
	Subject     string // optional current or desired subject
	CurrentHTML string // optional email to revise
	Tone        string // e.g. friendly, professional; empty lets the model choose
	Language    string // code (en, vi, zh, ja) or English language name
}

// Output is the generated email.
type Output struct {
	Subject      string `json:"subject"`
	Preheader    string `json:"preheader"`
	HTML         string `json:"html"`
	Model        string `json:"model"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

// Generator writes emails with an llm.Provider.
type Generator struct {
	provider  llm.Provider
	model     string
	maxTokens int
}

// Option customizes a Generator.
type Option func(*Generator)

// WithModel overrides the provider's default model.
func WithModel(model string) Option { return func(g *Generator) { g.model = model } }

// WithMaxTokens caps the reply length.
func WithMaxTokens(n int) Option { return func(g *Generator) { g.maxTokens = n } }

// NewGenerator returns a Generator backed by provider.
func NewGenerator(provider llm.Provider, opts ...Option) *Generator {
	g := &Generator{provider: provider, maxTokens: defaultMaxTokens}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// Generate writes a new email, or revises Input.CurrentHTML. A reply without a usable HTML
// document is retried once with a repair instruction.
func (g *Generator) Generate(ctx context.Context, in Input) (*Output, error) {
	in.Brief = strings.TrimSpace(in.Brief)
	if in.Brief == "" {
		return nil, ErrEmptyBrief
	}

	req := &llm.GenerateRequest{
		Model:       g.model,
		System:      buildSystemPrompt(in.Language),
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: buildUserPrompt(in)}},
		MaxTokens:   g.maxTokens,
		Temperature: llm.Float64(0.7),
	}

	out := &Output{}
	for attempt := 0; attempt < 2; attempt++ {
		resp, err := g.provider.Generate(ctx, req)
		if err != nil {
			return nil, err
		}
		out.InputTokens += resp.Usage.InputTokens
		out.OutputTokens += resp.Usage.OutputTokens
		out.Model = resp.Model

		subject, preheader, html, ok := parseReply(resp.Content)
		if ok {
			out.Subject, out.Preheader, out.HTML = subject, preheader, sanitizeHTML(html)
			if out.Subject == "" {
				out.Subject = strings.TrimSpace(in.Subject)
			}
			return out, nil
		}
		if resp.FinishReason == llm.FinishLength || attempt == 1 {
			break
		}
		req.Messages = append(req.Messages,
			llm.Message{Role: llm.RoleAssistant, Content: resp.Content},
			llm.Message{Role: llm.RoleUser, Content: repairPrompt},
		)
	}
	return nil, ErrInvalidModelOutput
}
