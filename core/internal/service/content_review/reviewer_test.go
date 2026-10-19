package content_review

import (
	"billionmail-core/internal/service/llm"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeProvider replies with queued contents and records every request it receives.
type fakeProvider struct {
	replies  []string
	err      error
	requests []llm.GenerateRequest
}

func (f *fakeProvider) Name() string { return "fake" }

func (f *fakeProvider) Generate(_ context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	cp := *req
	cp.Messages = append([]llm.Message(nil), req.Messages...)
	f.requests = append(f.requests, cp)
	if f.err != nil {
		return nil, f.err
	}
	reply := f.replies[0]
	f.replies = f.replies[1:]
	return &llm.GenerateResponse{
		Content: reply,
		Model:   "fake-model",
		Usage:   llm.Usage{InputTokens: 100, OutputTokens: 20},
	}, nil
}

const sampleContent = `<h1>Huge SALE!!!</h1><p>Earn money fast, 100% guaranteed.</p><a href="https://bit.ly/x">Claim</a>`

func TestReview(t *testing.T) {
	fake := &fakeProvider{replies: []string{`{
		"summary": "Có nhiều cụm từ spam.",
		"issues": [
			{"category":"misleading_claim","severity":"high","location":"body","excerpt":"Earn money fast, 100% guaranteed","reason":"Cam kết thu nhập","suggestion":"Bỏ cam kết"},
			{"category":"link_risk","severity":"medium","location":"link","excerpt":"https://bit.ly/x","reason":"Link rút gọn","suggestion":"Dùng URL đầy đủ"}
		]
	}`}}
	r := NewReviewer(fake, WithModel("custom-model"), WithMaxTokens(1000))

	res, err := r.Review(context.Background(), Input{
		Subject: "Huge sale", Content: sampleContent, Language: "vi", UnsubscribeManaged: true,
	})
	require.NoError(t, err)

	require.Len(t, fake.requests, 1)
	req := fake.requests[0]
	assert.Equal(t, "custom-model", req.Model)
	assert.Equal(t, 1000, req.MaxTokens)
	require.NotNil(t, req.Temperature)
	assert.Zero(t, *req.Temperature, "deterministic output")
	assert.Contains(t, req.System, "in Vietnamese.")
	assert.Contains(t, req.System, "appends an unsubscribe link")
	require.Len(t, req.Messages, 1)
	assert.Contains(t, req.Messages[0].Content, "<subject>Huge sale</subject>")
	assert.Contains(t, req.Messages[0].Content, `url: https://bit.ly/x`)

	assert.Equal(t, 5.5, res.Score)
	assert.Equal(t, VerdictReview, res.Verdict)
	assert.Equal(t, "Có nhiều cụm từ spam.", res.Summary)
	require.Len(t, res.Issues, 2)
	assert.Equal(t, LocationBody, res.Issues[0].Location)
	assert.Equal(t, LocationLink, res.Issues[1].Location)
	assert.Equal(t, "fake-model", res.Model)
	assert.Equal(t, PromptVersion, res.PromptVersion)
	assert.Equal(t, 100, res.InputTokens)
	assert.Equal(t, 20, res.OutputTokens)
}

func TestReview_CleanEmail(t *testing.T) {
	fake := &fakeProvider{replies: []string{`{"summary":"Looks fine.","issues":[]}`}}
	res, err := NewReviewer(fake).Review(context.Background(), Input{Subject: "Monthly update", Content: "<p>Hi there</p>"})
	require.NoError(t, err)
	assert.Equal(t, 10.0, res.Score)
	assert.Equal(t, VerdictPass, res.Verdict)
	assert.Empty(t, res.Issues)
	assert.NotNil(t, res.Issues, "serializes as [] rather than null")
}

func TestReview_RepairsInvalidJSONOnce(t *testing.T) {
	fake := &fakeProvider{replies: []string{
		"Sure! The email looks risky because...",
		`{"summary":"ok","issues":[]}`,
	}}
	res, err := NewReviewer(fake).Review(context.Background(), Input{Subject: "s", Content: "c"})
	require.NoError(t, err)
	assert.Equal(t, 10.0, res.Score)

	require.Len(t, fake.requests, 2)
	retry := fake.requests[1].Messages
	require.Len(t, retry, 3)
	assert.Equal(t, llm.RoleAssistant, retry[1].Role)
	assert.Equal(t, "Sure! The email looks risky because...", retry[1].Content)
	assert.Equal(t, repairPrompt, retry[2].Content)
	assert.Equal(t, 200, res.InputTokens, "usage covers both calls")
}

func TestReview_GivesUpAfterSecondInvalidReply(t *testing.T) {
	fake := &fakeProvider{replies: []string{"nope", "still nope"}}
	_, err := NewReviewer(fake).Review(context.Background(), Input{Subject: "s", Content: "c"})
	assert.ErrorIs(t, err, ErrInvalidModelOutput)
	assert.Len(t, fake.requests, 2)
}

func TestReview_Errors(t *testing.T) {
	_, err := NewReviewer(&fakeProvider{}).Review(context.Background(), Input{Subject: "  ", Content: "\n"})
	assert.ErrorIs(t, err, ErrEmptyInput)

	providerErr := &llm.APIError{Provider: "fake", StatusCode: 429, Message: "slow down"}
	fake := &fakeProvider{err: providerErr}
	_, err = NewReviewer(fake).Review(context.Background(), Input{Subject: "s", Content: "c"})
	assert.True(t, errors.Is(err, providerErr), "provider errors are returned unchanged")
	assert.Len(t, fake.requests, 1, "provider errors are not retried here; wrap the provider with llm.WithRetry")
}
